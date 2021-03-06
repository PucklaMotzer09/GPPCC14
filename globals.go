package main

import (
	"github.com/PucklaMotzer09/GoHomeEngine/src/gohome"
	"github.com/PucklaMotzer09/GoHomeEngine/src/physics2d"
)

var PhysicsMgr physics2d.PhysicsManager2D

const GAME_WIDTH = 1280
const GAME_HEIGHT = 720

const GRAVITY = 200
const ZOOM = 3

const PLAYER_DEPTH = 1
const DELETE_RAY_DEPTH = 2
const WEAPON_DEPTH = 3
const SPECIAL_DEPTH = 4
const INVENTORY_DEPTH = 5
const SCOPE_DEPTH = 6
const MENU_DEPTH = 7

const NUM_LEVELS = 9

var (
	LEVELS_TMX_MAPS = [NUM_LEVELS]string{
		"level1.tmx",
		"level2.tmx",
		"level3.tmx",
		"level4.tmx",
		"level5.tmx",
		"level6.tmx",
		"level7.tmx",
		"level8.tmx",
		"level9.tmx",
	}
)

var Camera gohome.Camera2D

const KEY_RIGHT = gohome.KeyD
const KEY_LEFT = gohome.KeyA
const KEY_DOWN = gohome.KeyS
const KEY_JUMP = gohome.KeyW
const KEY_JUMP1 = gohome.KeySpace
const KEY_SHOOT = gohome.MouseButtonLeft

const CAMERA_BOX_WIDTH float32 = float32(GAME_WIDTH) / ZOOM
const CAMERA_BOX_HEIGHT float32 = float32(GAME_HEIGHT) / ZOOM
const CAMERA_SPEED float32 = 0.1

var CAMERA_OFFSET = [2]float32{0.0, 0.0}

const GROUND_FRICTION float64 = 1.8
const PLAYER_CATEGORY uint16 = 1 << 0
const PLAYER_FEET_CATEGORY uint16 = (1 << 1) | PLAYER_CATEGORY
const GROUND_CATEGORY uint16 = 1 << 2
const WEAPON_CATEGORY uint16 = 1 << 3
const ENEMY_CATEGORY uint16 = 1 << 4
const ENEMY_SENSOR_CATEGORY uint16 = 1 << 5
const ENEMY_SMALL_LEFT_SENSOR_CATEGORY uint16 = 1<<6 | ENEMY_SENSOR_CATEGORY
const ENEMY_SMALL_RIGHT_SENSOR_CATEGORY uint16 = 1<<7 | ENEMY_SENSOR_CATEGORY
const ENEMY_BIG_LEFT_SENSOR_CATEGORY uint16 = 1<<8 | ENEMY_SENSOR_CATEGORY
const ENEMY_BIG_RIGHT_SENSOR_CATEGORY uint16 = 1<<9 | ENEMY_SENSOR_CATEGORY
const PLAYER_FEET_SENSOR_CATEGORY uint16 = (1 << 10) | PLAYER_CATEGORY
const SPIKE_CATEGORY uint16 = 1 << 11
const BALL_CATEGORY uint16 = 1 << 12

const AI_DISTANCE float32 = CAMERA_BOX_WIDTH * ZOOM / 3.0

const WIN_CONDITION_TARGET = true
const WIN_CONDITION_ENEMY = false

var CURRENT_WIN_CONDITION bool

func LoadResources() {
	gohome.ResourceMgr.LoadFont("Button", "/usr/share/fonts/truetype/ubuntu/UbuntuMono-R.ttf")
	gohome.ResourceMgr.LoadTexture("Player", "assets/textures/GPPCC14_Player.png")
	gohome.ResourceMgr.LoadTexture("DefaultWeapon", "assets/textures/GPPCC14_DefaultWeapon.png")
	gohome.ResourceMgr.LoadTexture("FreezeWeapon", "assets/textures/GPPCC14_FreezeWeapon.png")
	gohome.ResourceMgr.LoadTexture("BallWeapon", "assets/textures/GPPCC14_BallWeapon.png")
	gohome.ResourceMgr.LoadTexture("MoveWeapon", "assets/textures/GPPCC14_MoveWeapon.png")
	gohome.ResourceMgr.LoadTexture("DeleteWeapon", "assets/textures/GPPCC14_DeleteWeapon.png")
	gohome.ResourceMgr.LoadTexture("DefaultWeaponInv", "assets/textures/GPPCC14_DefaultWeaponInv.png")
	gohome.ResourceMgr.LoadTexture("FreezeWeaponInv", "assets/textures/GPPCC14_FreezeWeaponInv.png")
	gohome.ResourceMgr.LoadTexture("BallWeaponInv", "assets/textures/GPPCC14_BallWeaponInv.png")
	gohome.ResourceMgr.LoadTexture("MoveWeaponInv", "assets/textures/GPPCC14_MoveWeaponInv.png")
	gohome.ResourceMgr.LoadTexture("DeleteWeaponInv", "assets/textures/GPPCC14_DeleteWeaponInv.png")
	gohome.ResourceMgr.LoadTexture("DefaultWeaponBlock", "assets/textures/GPPCC14_DefaultWeaponBlock.png")
	gohome.ResourceMgr.LoadTexture("FreezeWeaponBlock", "assets/textures/GPPCC14_FreezeWeaponBlock.png")
	gohome.ResourceMgr.LoadTexture("BallWeaponBlock", "assets/textures/GPPCC14_BallWeaponBlock.png")
	gohome.ResourceMgr.LoadTexture("MoveWeaponBlock", "assets/textures/GPPCC14_MoveWeaponBlock.png")
	gohome.ResourceMgr.LoadTexture("Enemy", "assets/textures/GPPCC14_Enemy.png")
	gohome.ResourceMgr.LoadTexture("Explosion", "assets/textures/GPPCC14_Explosion.png")
	gohome.ResourceMgr.LoadTexture("Disappear", "assets/textures/GPPCC14_Disappear.png")
	gohome.ResourceMgr.LoadTexture("Restart", "assets/textures/GPPCC14_Restart.png")
	gohome.ResourceMgr.LoadTexture("Back", "assets/textures/GPPCC14_Back.png")
	gohome.ResourceMgr.LoadTexture("Pause", "assets/textures/GPPCC14_Pause.png")
	gohome.ResourceMgr.LoadTexture("Resume", "assets/textures/GPPCC14_Resume.png")
	gohome.ResourceMgr.LoadTexture("LevelButton1", "assets/textures/GPPCC14_LevelButton1.png")
	gohome.ResourceMgr.LoadTexture("LevelButtonPressed", "assets/textures/GPPCC14_LevelButtonPressed.png")
	gohome.ResourceMgr.LoadTexture("AmmoFont", "assets/textures/GPPCC14_AmmoFont.png")
	gohome.ResourceMgr.LoadTexture("Target", "assets/textures/GPPCC14_Target.png")
	gohome.ResourceMgr.LoadTexture("TargetCollect", "assets/textures/GPPCC14_TargetCollect.png")
	gohome.ResourceMgr.LoadTexture("Continue", "assets/textures/GPPCC14_Continue.png")
	gohome.ResourceMgr.LoadTexture("Scope", "assets/textures/GPPCC14_Scope.png")
	gohome.ResourceMgr.LoadSound("Jump", "assets/sounds/GPPCC14_Jump.wav")
	gohome.ResourceMgr.LoadSound("Shoot", "assets/sounds/GPPCC14_Shoot.wav")
	gohome.ResourceMgr.LoadSound("Explosion", "assets/sounds/GPPCC14_Explosion.wav")
	gohome.ResourceMgr.LoadSound("TargetCollect", "assets/sounds/GPPCC14_TargetCollect.wav")
	gohome.ResourceMgr.LoadSound("Button", "assets/sounds/GPPCC14_Button.wav")
	gohome.ResourceMgr.LoadSound("ButtonPressed", "assets/sounds/GPPCC14_ButtonPressed.wav")
	gohome.ResourceMgr.LoadTexture("Options", "assets/textures/GPPCC14_Options.png")

	gohome.ResourceMgr.GetTexture("Player").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("DefaultWeapon").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("FreezeWeapon").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("BallWeapon").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("MoveWeapon").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("DeleteWeapon").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("DefaultWeaponInv").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("FreezeWeaponInv").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("BallWeaponInv").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("MoveWeaponInv").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("DeleteWeaponInv").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("DefaultWeaponBlock").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("FreezeWeaponBlock").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("BallWeaponBlock").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("MoveWeaponBlock").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Enemy").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Explosion").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Disappear").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Restart").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Back").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Pause").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Resume").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("LevelButton1").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("LevelButtonPressed").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("AmmoFont").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Target").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("TargetCollect").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Continue").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Scope").SetFiltering(gohome.FILTERING_NEAREST)
	gohome.ResourceMgr.GetTexture("Options").SetFiltering(gohome.FILTERING_NEAREST)
}
