package common

type Broker struct {
	beforeHooks        map[ActionName][]ActionHook
	afterHooks         map[ActionName][]ActionHook
	runningAnimations  []Animation
	movementAnimations map[*Actor]*MovementAnimation
	queuedActions      [][]Action
	context            ActionContext
}


// NewBroker creates a new broker
// AND will inject a reference to itself into the context
func NewBroker(context ActionContext) *Broker {
	b := &Broker{
		beforeHooks:        make(map[ActionName][]ActionHook),
		afterHooks:         make(map[ActionName][]ActionHook),
		runningAnimations:  make([]Animation, 0),
		movementAnimations: make(map[*Actor]*MovementAnimation),
		queuedActions:      make([][]Action, 0),
	}
	context.ActionBroker = b
	b.context = context
	return b
}
func (d *Broker) Tick() {
	if len(d.movementAnimations) > 0 {
		for actor, animation := range d.movementAnimations {
			if animation.IsDone() {
				delete(d.movementAnimations, actor)
			} else {
				animation.Tick()
			}
		}
	}
	for i := len(d.runningAnimations) - 1; i >= 0; i-- {
		if d.runningAnimations[i].IsDone() {
			d.runningAnimations = append(d.runningAnimations[:i], d.runningAnimations[i+1:]...)
		} else {
			d.runningAnimations[i].Tick()
		}
	}

	d.checkForQueuedActions()
}

func (d *Broker) checkForQueuedActions() {
	animationsRunning := len(d.runningAnimations) > 0 || len(d.movementAnimations) > 0
	if !animationsRunning && len(d.queuedActions) > 0 {
		actions := d.queuedActions[0]
		d.queuedActions = d.queuedActions[1:]
		for _, action := range actions {
			d.Execute(action)
		}
	}
}

func (d *Broker) Execute(action Action) {
	if action == nil {
		return
	}

	var chainedActions []Action
	d.callBeforeHooks(action)

	chainedActions = action.Execute(d.context)
	animation := action.GetAnimations()
	if animation != nil {
		d.startAnimations(animation)
	}

	for _, chainedAction := range chainedActions {
		d.Execute(chainedAction)
	}

	d.callAfterHooks(action)

	return
}

func (d *Broker) QueueForExecutionAtTheSameTime(actions []Action) {
	if len(actions) == 0 {
		return
	}
	d.queuedActions = append(d.queuedActions, actions)
	d.checkForQueuedActions()

}
func (d *Broker) QueueForExecution(action Action) {
	d.QueueForExecutionAtTheSameTime([]Action{action})
}
func (d *Broker) Draw(renderer TileRenderer) {
	for _, animation := range d.movementAnimations {
		animation.Draw(renderer)
	}
	for _, animation := range d.runningAnimations {
		animation.Draw(renderer)
	}
}

func (d *Broker) callAfterHooks(action Action) {
	if hooks, ok := d.afterHooks[action.Name()]; ok {
		for _, hook := range hooks {
			chained := hook.After(action, d.context)
			for _, chainedAction := range chained {
				d.Execute(chainedAction)
			}
			//println(fmt.Sprintf("(%s) After Hook %s executed", action.Action(), hook.ActionName()))
		}
	}
	//println(fmt.Sprintf("(%s) Action executed", action.Action()))
}

func (d *Broker) callBeforeHooks(action Action) {
	if hooks, ok := d.beforeHooks[action.Name()]; ok {
		for _, hook := range hooks {
			chained := hook.Before(action, d.context)
			for _, chainedAction := range chained {
				d.Execute(chainedAction)
			}

			//println(fmt.Sprintf("(%s) Before Hook %s executed", action.Action(), hook.ActionName()))
		}
	}
}

func (d *Broker) RegisterHook(hook ActionHook) {
	if hook.IsAfter() {
		d.registerHook(hook, d.afterHooks)
	}
	if hook.IsBefore() {
		d.registerHook(hook, d.beforeHooks)
	}
}

func (d *Broker) registerHook(hook ActionHook, hooks map[ActionName][]ActionHook) {
	if _, ok := hooks[hook.ActionName()]; !ok {
		hooks[hook.ActionName()] = make([]ActionHook, 0)
	}
	hooks[hook.ActionName()] = append(hooks[hook.ActionName()], hook)
}

func (d *Broker) RemoveBeforeHook(hook ActionHook) {
	d.removeHook(hook, d.beforeHooks)
}

func (d *Broker) RemoveAfterHook(hook ActionHook) {
	d.removeHook(hook, d.afterHooks)
}
func (d *Broker) removeHook(hook ActionHook, hooks map[ActionName][]ActionHook) {
	if _, ok := hooks[hook.ActionName()]; ok {
		for i, h := range hooks[hook.ActionName()] {
			if h == hook {
				hooks[hook.ActionName()] = append(hooks[hook.ActionName()][:i], hooks[hook.ActionName()][i+1:]...)
				return
			}
		}
	}
}

func (d *Broker) AddAndRunAnimation(animation Animation) {
	d.runningAnimations = append(d.runningAnimations, animation)
	if animation.AudioCue() != "" {
		d.context.AudioPlayer.PlayCue(animation.AudioCue())
	}
}
func (d *Broker) startAnimations(animations []Animation) {
	for _, animation := range animations {
		switch typedAnimation := animation.(type) {
		case *MovementAnimation:
			d.StartMovementAnimation(typedAnimation)
		default:
			d.AddAndRunAnimation(animation)
		}
	}
}

func (d *Broker) StartMovementAnimation(newAnimation *MovementAnimation) {
	// idea: we will replace the existing with one that moves further
	if oldAnimation, ok := d.movementAnimations[newAnimation.GetActor()]; ok {
		if newAnimation.IsForThePlayer() {
			oldAnimation.Cancel()
			d.movementAnimations[newAnimation.GetActor()] = newAnimation
		} else {
			mergedAnimation := oldAnimation.Merge(newAnimation)
			d.movementAnimations[newAnimation.GetActor()] = mergedAnimation
		}
	} else {
		d.movementAnimations[newAnimation.GetActor()] = newAnimation
	}
	if newAnimation.AudioCue() != "" {
		d.context.AudioPlayer.PlayCue(newAnimation.AudioCue())
	}
}

func (d *Broker) GetAnimation(filter func(anim Animation) bool) (Animation, bool) {
	for _, animation := range d.runningAnimations {
		if filter(animation) {
			return animation, true
		}
	}
	for _, animation := range d.movementAnimations {
		if filter(animation) {
			return animation, true
		}
	}
	return nil, false
}

func (d *Broker) CancelAnimationsByFilter(filter func(anim Animation) bool) {
	for i := len(d.runningAnimations) - 1; i >= 0; i-- {
		animation := d.runningAnimations[i]
		if filter(animation) {
			animation.Cancel()
			d.runningAnimations = append(d.runningAnimations[:i], d.runningAnimations[i+1:]...)
		}
	}
}

func (d *Broker) CancelAllAnimations() {
	for _, animation := range d.runningAnimations {
		animation.Cancel()
	}
	d.runningAnimations = make([]Animation, 0)
	for _, animation := range d.movementAnimations {
		animation.Cancel()
	}
	d.movementAnimations = make(map[*Actor]*MovementAnimation)
}

func (d *Broker) ClearQueuedActions(shouldBeRemoved func(action Action) bool) {
	for i := len(d.queuedActions) - 1; i >= 0; i-- {
		actions := d.queuedActions[i]
		//for _, action := range actions {
		for v := len(actions) - 1; v >= 0; v-- {
			action := actions[v]
			if shouldBeRemoved(action) {
				if v == len(actions)-1 {
					actions = actions[:v]
				} else {
					actions = append(actions[:v], actions[v+1:]...)
				}
			}
		}
		if len(actions) == 0 {
			if i == len(d.queuedActions)-1 {
				d.queuedActions = d.queuedActions[:i]
			} else {
				d.queuedActions = append(d.queuedActions[:i], d.queuedActions[i+1:]...)
			}
		} else {
			d.queuedActions[i] = actions
		}
	}
}
