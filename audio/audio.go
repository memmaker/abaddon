package audio

import (
	"io"
	"io/fs"
	"math/rand"
	"path"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

func Clamp(value float64, minVal float64, maxVal float64) float64 {
	return min(max(value, minVal), maxVal)
}

type Cue struct {
	AudioCues    []string
	IsRandom     bool
	CurrentIndex int
}

func (c *Cue) Next() string {
	if c.IsRandom {
		randomIndex := rand.Intn(len(c.AudioCues))
		return c.AudioCues[randomIndex]
	}
	if c.CurrentIndex >= len(c.AudioCues) {
		c.CurrentIndex = 0
	}
	c.CurrentIndex++
	return c.AudioCues[c.CurrentIndex]
}

type Player struct {
	context         *audio.Context
	availableSounds map[string]*Sound
	activeSounds    map[string][]*ActiveSound
	queuedSounds    []string
	loopVolume      float64
	sfxVolume       float64
	masterVolume    float64
	files           fs.ReadDirFS
	disablePlayback bool
}

func (ap *Player) SetDisablePlayback(disable bool) {
	ap.disablePlayback = disable
}
func (ap *Player) GetMasterVolume() float64 {
	return ap.masterVolume
}

func (ap *Player) GetMusicVolume() float64 {
	return ap.loopVolume
}

func (ap *Player) GetSoundVolume() float64 {
	return ap.sfxVolume
}

func (ap *Player) effectiveMusicVolume() float64 {
	return ap.loopVolume * ap.masterVolume
}

func (ap *Player) effectiveSoundVolume() float64 {
	return ap.sfxVolume * ap.masterVolume
}

func (ap *Player) SetMasterVolume(volume float64) {
	ap.masterVolume = Clamp(volume, 0, 1)
	ap.updateVolume()
}

func (ap *Player) SetMusicVolume(volume float64) {
	ap.loopVolume = Clamp(volume, 0, 1)
	ap.updateVolume()
}

func (ap *Player) SetSoundVolume(volume float64) {
	ap.sfxVolume = Clamp(volume, 0, 1)
	ap.updateVolume()
}

func (ap *Player) updateVolume() {
	for cue, activeSounds := range ap.activeSounds {
		sound := ap.availableSounds[cue]
		for _, activeSound := range activeSounds {
			if sound.IsLooping {
				activeSound.player.SetVolume(ap.effectiveMusicVolume())
			} else {
				activeSound.player.SetVolume(ap.effectiveSoundVolume())
			}
		}
	}
}

type Sound struct {
	Cue          string
	Filenames    []string
	Samples      [][]byte
	IsLooping    bool
	IsRandomized bool
	currentIndex int
}

func (s *Sound) Unload() {
	s.Samples = make([][]byte, len(s.Filenames))
}
func (s *Sound) IsLoaded() bool {
	return len(s.Samples) > 0 && s.Samples[0] != nil && len(s.Samples[0]) > 0
}

func (s *Sound) Load(ap *Player) bool {
	for i, filename := range s.Filenames {
		//print(fmt.Sprintf("Loading %s...", filename))
		oggData, success := ap.LoadOGG(filename)
		if !success {
			println("ERR: Could not load sound file: " + filename)
			return false
		}
		s.Samples[i] = oggData
		//println("done")
	}
	if !s.IsLoaded() {
		println("ERR: Sound cue not found: " + s.Cue)
		return false
	}
	return true
}
func (s *Sound) Next() []byte {
	if s.IsRandomized {
		return s.Samples[rand.Intn(len(s.Samples))]
	}
	sample := s.Samples[s.currentIndex]
	s.currentIndex = (s.currentIndex + 1) % len(s.Samples)
	return sample
}

func (s *Sound) nextFilename() string {
	if s.IsRandomized {
		return s.Filenames[rand.Intn(len(s.Filenames))]
	}
	filename := s.Filenames[s.currentIndex]
	s.currentIndex = (s.currentIndex + 1) % len(s.Filenames)
	return filename
}

func (s *Sound) OpenStream(ap *Player) *audio.Player {
	return ap.OpenOGG(s.nextFilename())
}

type ActiveSound struct {
	player   *audio.Player
	callback func()
}

type NullHandle struct{}

func (h *NullHandle) Close() error    { return nil }
func (h *NullHandle) IsPlaying() bool { return false }

type SampleRate int

const (
	SampleRate22050 SampleRate = 22050
	SampleRate44100 SampleRate = 44100
)

func NewAudioPlayer(fileSystem fs.ReadDirFS, sampleRate SampleRate) *Player {
	return &Player{
		files:           fileSystem,
		context:         audio.NewContext(int(sampleRate)),
		availableSounds: make(map[string]*Sound),
		activeSounds:    make(map[string][]*ActiveSound),
		sfxVolume:       0.2,
		loopVolume:      1,
		masterVolume:    1,
	}
}
func (ap *Player) UnloadAll() {
	for _, sound := range ap.availableSounds {
		sound.Unload()
	}
	ap.activeSounds = make(map[string][]*ActiveSound)
}

func (ap *Player) StopAll() {
	for cue, sounds := range ap.activeSounds {
		for _, sound := range sounds {
			sound.player.Close()
		}
		delete(ap.activeSounds, cue)
	}
}

// LoadOGG loads an OGG file from the given filename into memory.
func (ap *Player) LoadOGG(filename string) ([]byte, bool) {
	file, filErr := ap.files.Open(filename)
	if filErr != nil {
		return nil, false
	}
	defer file.Close()
	stream, decodeErr := vorbis.DecodeWithoutResampling(file)
	if decodeErr != nil || filErr != nil {
		return nil, false
	}
	_, SeekErr := stream.Seek(0, io.SeekStart)
	if SeekErr != nil {
		return nil, false
	}
	buffer := make([]byte, stream.Length())
	_, err := stream.Read(buffer)
	if err != nil {
		return nil, false
	}
	return buffer, true
}

// OpenOGG opens an OGG file from the given filename for streaming.
func (ap *Player) OpenOGG(filename string) *audio.Player {
	file, filErr := ap.files.Open(filename)
	if filErr != nil {
		return nil
	}
	defer file.Close()
	stream, decodeErr := vorbis.DecodeWithoutResampling(file)
	if decodeErr != nil || filErr != nil {
		return nil
	}
	player, err := ap.context.NewPlayer(stream)
	if err != nil {
		return nil
	}
	return player
}

func (ap *Player) Stop(cue string) {
	if _, ok := ap.activeSounds[cue]; ok {
		for _, sound := range ap.activeSounds[cue] {
			sound.player.Close()
		}
		delete(ap.activeSounds, cue)
	}
}

type Handle interface {
	Close() error
	IsPlaying() bool
}

func (ap *Player) StartLoop(cue string) Handle {
	if ap.disablePlayback {
		return &NullHandle{}
	}
	if sound, ok := ap.availableSounds[cue]; ok {
		if !sound.IsLoaded() && !sound.Load(ap) {
			return &NullHandle{}
		}
		sound.IsLooping = true
		return ap.startPlaying(sound, ap.effectiveMusicVolume(), nil)
	} else {
		println("ERR: Sound cue not found: " + cue)
	}
	return &NullHandle{}
}

func (ap *Player) StartLoopStream(cue string) Handle {
	if ap.disablePlayback {
		return &NullHandle{}
	}
	if sound, ok := ap.availableSounds[cue]; ok {
		sound.IsLooping = true
		player := sound.OpenStream(ap)
		if player == nil {
			return &NullHandle{}
		}
		player.Rewind()
		player.SetVolume(ap.effectiveMusicVolume())
		player.Play()
		ap.addActiveSound(sound, player, nil)
		return player
	} else {
		println("ERR: Sound cue not found: " + cue)
	}
	return &NullHandle{}
}

func (ap *Player) startPlaying(sound *Sound, volume float64, callback func()) *audio.Player {
	player := ap.context.NewPlayerFromBytes(sound.Next())
	player.Rewind()
	player.SetVolume(volume)
	player.Play()
	ap.addActiveSound(sound, player, callback)
	return player
}
func (ap *Player) addActiveSound(sound *Sound, player *audio.Player, callback func()) {
	if _, ok := ap.activeSounds[sound.Cue]; !ok {
		ap.activeSounds[sound.Cue] = make([]*ActiveSound, 0)
	}
	ap.activeSounds[sound.Cue] = append(ap.activeSounds[sound.Cue], &ActiveSound{player: player, callback: callback})
}
func (ap *Player) PlayCue(cue string) Handle {
	if ap.disablePlayback {
		return &NullHandle{}
	}
	if sound, ok := ap.availableSounds[cue]; ok {
		if !sound.IsLoaded() && !sound.Load(ap) {
			return &NullHandle{}
		}
		return ap.startPlaying(sound, ap.effectiveSoundVolume(), nil)
	} else {
		println("ERR: Sound cue not found: " + cue)
	}
	return &NullHandle{}
}

func (ap *Player) PlayCueWhenFree(cue string) {
	if ap.disablePlayback {
		return
	}
	if !ap.IsSoundEffectPlaying() {
		ap.PlayCue(cue)
	}
	ap.queuedSounds = append(ap.queuedSounds, cue)
}

func (ap *Player) PlayCueIfFree(cue string) {
	if ap.disablePlayback {
		return
	}
	if !ap.IsSoundEffectPlaying() {
		ap.PlayCue(cue)
	}
}

func (ap *Player) PlayCueWithCallback(cue string, callback func()) Handle {
	if ap.disablePlayback {
		return &NullHandle{}
	}
	if sound, ok := ap.availableSounds[cue]; ok {
		if !sound.IsLoaded() && !sound.Load(ap) {
			return &NullHandle{}
		}
		return ap.startPlaying(sound, ap.effectiveSoundVolume(), callback)
	} else {
		println("ERR: Sound cue not found: " + cue)
	}
	return &NullHandle{}
}

func (ap *Player) IsCuePlaying(cue string) bool {
	if sounds, ok := ap.activeSounds[cue]; ok {
		return len(sounds) > 0
	}
	return false
}

func (ap *Player) Update() {
	if ap.disablePlayback {
		return
	}
	isSoundEffectPlaying := false
	for cue, sounds := range ap.activeSounds {
		soundDef := ap.availableSounds[cue]
		for i := len(sounds) - 1; i >= 0; i-- {
			sound := sounds[i]
			if !sound.player.IsPlaying() {
				if soundDef.IsLooping {
					sound.player.Rewind()
					sound.player.Play()
					continue
				}
				sound.player.Close()
				if sound.callback != nil {
					sound.callback()
				}
				ap.activeSounds[cue] = append(ap.activeSounds[cue][:i], ap.activeSounds[cue][i+1:]...)
			} else if !soundDef.IsLooping {
				isSoundEffectPlaying = true
			}
		}
	}
	if !isSoundEffectPlaying && len(ap.queuedSounds) > 0 {
		cue := ap.queuedSounds[0]
		ap.queuedSounds = ap.queuedSounds[1:]
		ap.PlayCue(cue)
	}
}
func (ap *Player) RegisterSoundCues(filenames []string) {
	files := ap.files
	for _, filename := range filenames {
		file, err := files.Open(filename)
		if err != nil {
			println("ERR: Sound file not found: " + filename)
			continue
		}
		file.Close()
		extension := strings.ToLower(filepath.Ext(filename))
		basename := filepath.Base(filename)
		cue := basename[0 : len(basename)-len(extension)]
		s := &Sound{
			Cue:       cue,
			Filenames: []string{filename},
			Samples:   make([][]byte, 1),
		}
		ap.availableSounds[cue] = s
	}
}
func (ap *Player) PreLoadCuesIntoMemory(soundCues []string) {
	for _, soundCue := range soundCues {
		if sound, ok := ap.availableSounds[soundCue]; ok {
			if !sound.IsLoaded() {
				//println(fmt.Sprintf("Loading sound cue '%s'", soundCue))
				sound.Load(ap)
			}
		}
	}
}
func (ap *Player) UnloadCues(soundCues []string) {
	for _, soundCue := range soundCues {
		if sound, ok := ap.availableSounds[soundCue]; ok {
			if sound.IsLoaded() {
				//println(fmt.Sprintf("Unloading sound cue '%s'", soundCue))
				sound.Unload()
			}
		}
	}
}
func (ap *Player) RegisterRandomizedSoundCues(directories []string) {
	for _, directory := range directories {
		cue := filepath.Base(directory)
		filesInDirectory, _ := ap.files.ReadDir(directory)
		var fileNames []string
		for _, file := range filesInDirectory {
			filename := file.Name()
			if !strings.HasSuffix(filename, ".ogg") {
				continue
			}
			fileNames = append(fileNames, path.Join(directory, filename))
		}
		s := &Sound{
			Cue:          cue,
			Filenames:    fileNames,
			Samples:      make([][]byte, len(filesInDirectory)),
			IsRandomized: true,
		}
		ap.availableSounds[cue] = s
	}
}

func (ap *Player) IsSoundEffectPlaying() bool {
	for cue, sounds := range ap.activeSounds {
		for _, sound := range sounds {
			soundDef := ap.availableSounds[cue]
			if sound.player.IsPlaying() && !soundDef.IsLooping {
				return true
			}
		}
	}
	return false
}
