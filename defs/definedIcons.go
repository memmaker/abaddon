package defs

/*
- Grass
- Dirt
- Mountain
- MountainWithEntrance
- Water
- WoodenBridge
- Tree
- DungeonFloor
- DungeonWall
- DungeonWallWithShading
- OpenDoor
- ClosedDoor
- StairsUp
- StairsDown

Effects
- Blood
- Fire
- Lightning
- HitEffects

Items
- Potions
- Scrolls
- Common, Magical, Unique Items
  - Swords
  - Clubs
  - Axes
  - Staves
  - Bows
  - Armor
  - Shields

Characters
  - Rogue
  - Warrior
  - Sorcerer

Monsters
  - Fallen
  - Goat Man
  - Goat Man Archer
  - Hidden
  - Scavenger
  - Skeleton
  - Skeleton Archer
  - Skeleton Captain
  - Winged Fiend
  - Zombie
  - Gargoyle
  - Horned Demon
  - Magma Demon
  - Overlord
  - Spitting Terror
  - Balrog
  - Hell Knight
  - Lightning Demon
  - Succubus
  - Viper
  - Mage

Unique Monsters
  - The Butcher
  - Skeleton King
  - Gharbad the Weak
  - Zhar the Mad
  - Warlord of Blood
  - Archbishop Lazarus
  - Black Jade
  - Red Vex
  - Snotspill
  - Lachdanan
  - Blackjade
*/
const (
	IconNone                     int32 = -1
	IconWhite                    int32 = 18
	IconTargeting                int32 = 220
	IconGrass                    int32 = 0
	IconDirt                     int32 = 1
	IconMountain                 int32 = 7
	IconMountainWithEntrance     int32 = 8
	IconWater                    int32 = 28
	IconSpikes                   int32 = 264
	IconLava                     int32 = 159
	IconWoodenBridge             int32 = 63
	IconTreeDark                 int32 = 5
	IconTreeLight                int32 = 4
	IconDungeonOfDoomFloorModern int32 = 157
	IconDungeonOfDoomWallModern  int32 = 158
	IconDungeonOfDoomDoorModern  int32 = 241

	IconDungeonOfDoomFloor         int32 = 242
	IconDungeonOfDoomWallHoriz     int32 = 269
	IconDungeonOfDoomWallVert      int32 = 297
	IconDungeonOfDoomDoor          int32 = 240
	IconDungeonOfDoomCorridorWall  int32 = 243
	IconDungeonOfDoomCorridorFloor int32 = 268

	IconChasm                    int32 = 265
	IconCathedralFloor           int32 = 64
	IconCatacombsFloor           int32 = 86
	IconCatacombsWall            int32 = 84
	IconCatacombsWallWithShading int32 = 85
	IconDungeonFloorAlternative1 int32 = 14
	IconDungeonRedEarthFloor     int32 = 186
	IconDungeonRockyWall         int32 = 9
	IconCathedralWall            int32 = 56
	IconCathedralWallWithShading int32 = 58
	IconOpenDoor                 int32 = 70
	IconClosedDoor               int32 = 69
	IconStairsUp                 int32 = 74
	IconStairsDown               int32 = 73
	IconShrine                   int32 = 75
	IconSign                     int32 = 154

	IconDecalBlood1         int32 = 189
	IconDecalBlood2         int32 = 190
	IconDecalBlood3         int32 = 191
	IconDecalBlood4         int32 = 192
	IconDecalBlood5         int32 = 193
	IconEffectLightning     int32 = 37
	IconEffectFireSmall     int32 = 109
	IconEffectFireBig       int32 = 32
	IconBarrel              int32 = 125
	IconEffectHit           int32 = 43
	IconItemPotionCrimson   int32 = 161
	IconItemPotionRed       int32 = 332
	IconItemPotionBlue      int32 = 329
	IconItemPotionDarkBlue  int32 = 327
	IconItemPotionGreen     int32 = 330
	IconItemPotionDarkGreen int32 = 335
	IconItemPotionYellow    int32 = 334
	IconItemPotionOrange    int32 = 333
	IconItemPotionPink      int32 = 331
	IconItemPotionBrown     int32 = 326
	IconItemPotionWhite     int32 = 328
	IconItemFlaskOfOil      int32 = 246

	IconCloakDark        int32 = 411
	IconGlovesDark       int32 = 439
	IconBootsDark        int32 = 440
	IconArmorDark        int32 = 441
	IconHoodDark         int32 = 467
	IconRingDark         int32 = 468
	IconHelmetDark       int32 = 469
	IconShieldDarkThorns int32 = 495
	IconAmuletDark       int32 = 496
	IconScytheDark       int32 = 497
	IconMalletGolden     int32 = 498

	IconItemScroll    int32 = 139
	IconItemTorch     int32 = 352
	IconItemLantern   int32 = 247
	IconItemStarGlass int32 = 248

	IconItemMorningStar int32 = 274

	IconItemRing1   int32 = 129
	IconItemRing2   int32 = 130
	IconItemRing3   int32 = 131
	IconItemAmulet1 int32 = 132
	IconItemAmulet2 int32 = 133
	IconItemAmulet3 int32 = 134

	IconItemSword       int32 = 144
	IconItemDagger      int32 = 233
	IconItemMagicSword  int32 = 146
	IconItemUniqueSword int32 = 147

	IconItemMace       int32 = 168
	IconItemMagicMace  int32 = 169
	IconItemUniqueMace int32 = 170

	IconItemAxe       int32 = 171
	IconItemMagicAxe  int32 = 172
	IconItemUniqueAxe int32 = 173

	IconItemStaff       int32 = 141
	IconItemMagicStaff  int32 = 143
	IconItemUniqueStaff int32 = 140
	IconItemWand        int32 = 124
	IconItemBow         int32 = 149
	IconItemMagicBow    int32 = 150
	IconItemUniqueBow   int32 = 151

	IconItemDart int32 = 232

	IconItemLeatherArmor1   int32 = 227
	IconItemLeatherArmor2   int32 = 228
	IconItemLeatherArmor3   int32 = 229
	IconItemChainMail       int32 = 196
	IconItemUniqueChainMail int32 = 197
	IconItemArmor           int32 = 199
	IconItemMagicArmor      int32 = 198
	IconItemUniqueArmor     int32 = 201
	IconItemHelmet          int32 = 211
	IconItemMagicHelmet     int32 = 212
	IconItemUniqueHelmet    int32 = 213

	IconItemBoots1 int32 = 355
	IconItemBoots2 int32 = 356
	IconItemBoots3 int32 = 357
	IconItemBoots4 int32 = 358
	IconItemBoots5 int32 = 359
	IconItemBoots6 int32 = 360
	IconItemBoots7 int32 = 361
	IconItemBoots8 int32 = 362
	IconItemBoots9 int32 = 363

	IconItemGloves1 int32 = 383
	IconItemGloves2 int32 = 384
	IconItemGloves3 int32 = 385
	IconItemGloves4 int32 = 386
	IconItemGloves5 int32 = 387
	IconItemGloves6 int32 = 388
	IconItemGloves7 int32 = 389
	IconItemGloves8 int32 = 390
	IconItemGloves9 int32 = 391

	IconItemCape1 int32 = 412
	IconItemCape2 int32 = 413
	IconItemCape3 int32 = 414
	IconItemCape4 int32 = 415
	IconItemCape5 int32 = 416
	IconItemCape6 int32 = 417
	IconItemCape7 int32 = 418
	IconItemCape8 int32 = 419

	IconItemShield       int32 = 205
	IconItemMagicShield  int32 = 206
	IconItemUniqueShield int32 = 207

	IconCharacterRogueMale      int32 = 19
	IconCharacterRogueFemale    int32 = 47
	IconCharacterWarriorMale    int32 = 20
	IconCharacterWarriorFemale  int32 = 48
	IconCharacterSorcererMale   int32 = 21
	IconCharacterSorcererFemale int32 = 49

	IconMonsterBat              int32 = 93  // level 1-3
	IconMonsterSkeleton         int32 = 397 // level 1-3
	IconMonsterSkeletonWarrior  int32 = 398 // level 2-4
	IconMonsterSkeletonArcher   int32 = 401 // level 2-6
	IconMonsterSkeletonSorcerer int32 = 402 // level 2-6
	IconMonsterSkeletonCaptain  int32 = 400 // level 2-6
	IconMonsterFallen           int32 = 426 // level 1-6
	IconMonsterFallenWarrior    int32 = 427 // level 5-8
	IconMonsterFallenArcher     int32 = 425 // level 5-8
	IconMonsterFallenShaman     int32 = 426 // level 5-8

	IconMonsterCaveling           int32 = 542
	IconMonsterCavelingWithWeapon int32 = 541
	IconMonsterCaveImpWithAxes    int32 = 540
	IconMonsterCaveImpWithBigAxe  int32 = 539
	IconMonsterCaveBear           int32 = 538

	IconMonsterFiend            int32 = 548
	IconMonsterHornedFiend      int32 = 547
	IconMonsterHornedWithKnife  int32 = 546
	IconMonsterHornedWithGlaive int32 = 545
	IconMonsterHornedWithWhip   int32 = 544
	IconMonsterHornedBig        int32 = 543

	IconMonsterThiefPossessed int32 = 463
	IconMonsterThiefGoblin    int32 = 491

	IconMonsterGoatMan       int32 = 95
	IconMonsterGoatManArcher int32 = 95
	IconMonsterHidden        int32 = 99
	IconMonsterScavenger     int32 = 96

	IconMonsterWingedFiend int32 = 105

	IconUniqueMeatJack  int32 = 258
	IconUniqueKingZedek int32 = 259
	IconUniqueRogue     int32 = 315
	IconUniqueRogueDog  int32 = 343
	IconUniqueLucifer   int32 = 261

	IconUniqueEvilShekel  int32 = 287
	IconUniqueEvilKirioth int32 = 288

	IconUniqueAdventurerKnight    int32 = 286
	IconUniqueAdventurerPriest    int32 = 314
	IconUniqueAdventurerBarbarian int32 = 342
	IconUniqueAdventurerArcher    int32 = 370

	IconCorpseFallen int32 = 135

	IconRogueBat     int32 = 567
	IconRogueEmu     int32 = 570
	IconRogueJackal  int32 = 575
	IconRogueKestrel int32 = 576
	IconRogueOrc     int32 = 580
	IconRogueSnake   int32 = 584

	IconCorpse      int32 = 135
	IconRogueCorpse int32 = 592
)
