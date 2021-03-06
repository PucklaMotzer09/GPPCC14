package main

import (
	"github.com/ByteArena/box2d"
	"github.com/PucklaMotzer09/GoHomeEngine/src/gohome"
	"github.com/PucklaMotzer09/GoHomeEngine/src/physics2d"
	"github.com/PucklaMotzer09/mathgl/mgl32"
	"strings"
)

const (
	DEATH_BUTTON_SIZE    float32 = 100.0
	DEATH_BUTTON_PADDING float32 = 100.0
	DEATH_TEXT_PADDING   float32 = 25.0

	PAUSE_BUTTON_SIZE float32 = 25.0
	PAUSE_BUTTON_X    float32 = float32(GAME_WIDTH) - PAUSE_BUTTON_SIZE/2.0 - PAUSE_BUTTON_SIZE/5.0
	PAUSE_BUTTON_Y    float32 = PAUSE_BUTTON_SIZE/2.0 + PAUSE_BUTTON_SIZE/5.0

	OPTIONS_BUTTON_SIZE float32 = PAUSE_BUTTON_SIZE
	OPTIONS_BUTTON_X    float32 = PAUSE_BUTTON_X - PAUSE_BUTTON_SIZE/2.0 - OPTIONS_BUTTON_SIZE
	OPTIONS_BUTTON_Y    float32 = PAUSE_BUTTON_Y
)

type LevelScene struct {
	LevelID        uint32
	Map            gohome.TiledMap
	Player         Player
	Enemies        []*Enemy
	Targets        []*Target
	targetCollects []*TargetCollect
	debugInfo      DebugInfo

	debugDraw physics2d.PhysicsDebugDraw2D

	deathBtns   [2]*gohome.Button
	winMenu     WinMenu
	optionsMenu OptionsMenu
	pauseBtn    *gohome.Button
	optionsBtn  *gohome.Button
	deathText   *gohome.Text2D
	levelTitle  LevelTitle

	menuInited    bool
	menuDirection bool
	paused        bool
	restarting    bool
}

func (this *LevelScene) Init() {
	gohome.ErrorMgr.ShowMessageBoxes = false
	if this.LevelID > NUM_LEVELS-1 {
		gohome.SceneMgr.SwitchScene(&LevelSelectScene{})
		return
	}

	physics2d.PIXEL_PER_METER = 10.0
	gohome.ResourceMgr.LoadTMXMap("Level", LEVELS_TMX_MAPS[this.LevelID])

	PhysicsMgr.Init([2]float32{0.0, GRAVITY})
	gohome.UpdateMgr.AddObject(&PhysicsMgr)
	this.debugDraw = PhysicsMgr.GetDebugDraw()
	this.debugDraw.Visible = false
	gohome.RenderMgr.AddObject(&this.debugDraw)

	this.initMap()
	this.initMenus()

	Camera.Position = [2]float32{-CAMERA_BOX_WIDTH, -CAMERA_BOX_HEIGHT}
}

func (this *LevelScene) initMap() {
	this.Map.Init("Level")
	gohome.RenderMgr.AddObject(&this.Map)
	groundBodies := PhysicsMgr.LayerToCollision(&this.Map, "Collision")
	for i := 0; i < len(groundBodies); i++ {
		b := groundBodies[i]
		if b == nil {
			continue
		}
		for f := b.GetFixtureList(); f != nil; f = f.GetNext() {
			filter := f.GetFilterData()
			filter.CategoryBits = GROUND_CATEGORY
			filter.MaskBits = 0xffff
			f.SetFilterData(filter)
			f.SetFriction(GROUND_FRICTION)
		}
	}

	var playerStart [2]float32

	ls := this.Map.Layers
	for i := 0; i < len(ls); i++ {
		l := ls[i]
		if l.Name == "Settings" {
			objs := l.Objects
			for j := 0; j < len(objs); j++ {
				o := objs[j]
				if o.Name == "start" {
					playerStart[0] = float32(o.X)
					playerStart[1] = float32(o.Y)
				} else if o.Name == "enemy" {
					enemy := &Enemy{}
					enemy.Sprite2D.Init("")
					enemy.Transform.Position = [2]float32{float32(o.X), float32(o.Y)}
					this.Enemies = append(this.Enemies, enemy)
				} else if o.Name == "target" {
					var target Target
					target.Init("Target")
					target.Transform.Origin = [2]float32{0.5, 0.5}
					target.Transform.Position = [2]float32{float32(o.X), float32(o.Y)}
					gohome.RenderMgr.AddObject(&target)
					this.Targets = append(this.Targets, &target)
				}
			}
		}
	}

	this.Player.Init(playerStart, &PhysicsMgr)
	for i := 0; i < len(this.Enemies); i++ {
		this.Enemies[i].Init(this.Enemies[i].Transform.Position, &this.Player)
	}

	CURRENT_WIN_CONDITION = WIN_CONDITION_TARGET
	mapprops := this.Map.Properties
	if mapprops != nil {
		props := mapprops.Properties
		for i := 0; i < len(props); i++ {
			p := props[i]
			if p.Name == "win_condition" {
				if p.Value == "enemy" {
					CURRENT_WIN_CONDITION = WIN_CONDITION_ENEMY
				} else if p.Value == "target" {
					CURRENT_WIN_CONDITION = WIN_CONDITION_TARGET
				}
			} else if strings.Contains(p.Name, "weapon") {
				if p.Name == "defaultweapon" && p.Value == "true" {
					this.Player.addWeapon(&DefaultWeapon{})
				} else if p.Name == "freezeweapon" && p.Value == "true" {
					this.Player.addWeapon(&FreezeWeapon{})
				} else if p.Name == "ballweapon" && p.Value == "true" {
					this.Player.addWeapon(&BallWeapon{})
				} else if p.Name == "moveweapon" && p.Value == "true" {
					this.Player.addWeapon(&MoveWeapon{})
				} else if p.Name == "deleteweapon" && p.Value == "true" {
					this.Player.addWeapon(&DeleteWeapon{})
				}
			}
		}
	}

	if len(this.Player.weapons) == 0 {
		this.Player.addWeapon(&DefaultWeapon{})
	}

	for _, l := range this.Map.Layers {
		if l.Data != nil {
			data := l.Data
			iter, err := data.Iter()
			if err != nil {
				gohome.ErrorMgr.Error("Level", "Spikes", "Couldn't get Iterator")
				break
			}
			for iter.Next() {
				tile := iter.Get()
				if tile.GID() >= 85 && tile.GID() <= 89 {
					counter := iter.GetIndex()
					pos := mgl32.Vec2{
						float32((counter % this.Map.Width) * this.Map.TileWidth),
						float32(((counter - (counter % this.Map.Width)) / this.Map.Width) * this.Map.TileHeight),
					}

					this.createSpike(pos)
				}
			}
		}
	}
}

func (this *LevelScene) createSpike(pos mgl32.Vec2) {
	w, h := float32(this.Map.TileWidth), float32(this.Map.TileHeight)

	bdef := box2d.MakeB2BodyDef()
	bdef.Type = box2d.B2BodyType.B2_staticBody
	bdef.Position = physics2d.ToBox2DCoordinates([2]float32{
		pos[0] + w/2.0,
		pos[1] + h/2.0,
	})

	body := PhysicsMgr.World.CreateBody(&bdef)

	fdef := box2d.MakeB2FixtureDef()
	fdef.Friction = GROUND_FRICTION
	fdef.Filter.CategoryBits = SPIKE_CATEGORY
	shape := box2d.MakeB2PolygonShape()
	shape.SetAsBox(physics2d.ScalarToBox2D(w/2.0), physics2d.ScalarToBox2D(h/2.0))
	fdef.Shape = &shape

	body.CreateFixtureFromDef(&fdef)
}

func (this *LevelScene) initMenus() {
	this.pauseBtn = &gohome.Button{}
	this.pauseBtn.Init([2]float32{PAUSE_BUTTON_X, PAUSE_BUTTON_Y}, "Pause")
	this.pauseBtn.Transform.Origin = [2]float32{0.5, 0.5}
	this.pauseBtn.Transform.Size = [2]float32{PAUSE_BUTTON_SIZE, PAUSE_BUTTON_SIZE}
	this.pauseBtn.Depth = MENU_DEPTH
	this.pauseBtn.PressCallback = func(btn *gohome.Button) {
		if this.winMenu.direction == DOWN {
			return
		}

		if this.paused {
			this.optionsMenu.direction = UP
			this.Resume()
		} else {
			this.Pause()
		}
	}

	this.optionsBtn = &gohome.Button{}
	this.optionsBtn.Init([2]float32{OPTIONS_BUTTON_X, OPTIONS_BUTTON_Y}, "Options")
	this.optionsBtn.Transform.Origin = [2]float32{0.5, 0.5}
	this.optionsBtn.Transform.Size = [2]float32{OPTIONS_BUTTON_SIZE, OPTIONS_BUTTON_SIZE}
	this.optionsBtn.Depth = MENU_DEPTH
	this.optionsBtn.PressCallback = func(btn *gohome.Button) {
		if this.winMenu.direction == DOWN || this.Player.Died() {
			return
		}

		this.optionsMenu.direction = !this.optionsMenu.direction
		if this.menuDirection == DOWN {
			this.menuDirection = UP
		}
		if this.optionsMenu.direction == DOWN {
			this.PauseGame()
		} else {
			this.Resume()
		}
	}

	this.winMenu.Init()
	this.optionsMenu.Init()
	this.debugInfo.Init()
	this.levelTitle.Level = uint8(this.LevelID + 1)
	this.levelTitle.WinCondition = CURRENT_WIN_CONDITION
	this.levelTitle.Init()
}

func (this *LevelScene) terminateMenu() {
	for _, btn := range this.deathBtns {
		if btn != nil {
			btn.Terminate()
		}
	}
	if this.deathText != nil {
		gohome.RenderMgr.RemoveObject(this.deathText)
		this.deathText.Terminate()
	}
	this.menuInited = false
}

func (this *LevelScene) initMenu(death bool, inMid bool) {
	if this.menuInited {
		return
	}

	if this.deathBtns[0] != nil {
		this.terminateMenu()
	}

	var restartBtn, backBtn gohome.Button

	width := 2.0*DEATH_BUTTON_SIZE + DEATH_BUTTON_PADDING
	mid := gohome.Render.GetNativeResolution().Mul(0.5)
	var restartPos, backPos, deathTextPos mgl32.Vec2

	if !inMid {
		restartPos = mid.Add([2]float32{
			-width/2.0 + DEATH_BUTTON_SIZE/2.0,
			-mid.Y() - DEATH_BUTTON_SIZE,
		})
		backPos = mid.Add([2]float32{
			width/2.0 - DEATH_BUTTON_SIZE/2.0,
			-mid.Y() - DEATH_BUTTON_SIZE,
		})
		deathTextPos = mid.Add([2]float32{
			10.0,
			-mid.Y() - DEATH_BUTTON_SIZE*1.5 - DEATH_TEXT_PADDING,
		})
		this.menuDirection = DOWN

	} else {
		restartPos = mid.Add([2]float32{
			-width/2.0 + DEATH_BUTTON_SIZE/2.0,
			0.0,
		})
		backPos = mid.Add([2]float32{
			width/2.0 - DEATH_BUTTON_SIZE/2.0,
			0.0,
		})
		deathTextPos = mid.Add([2]float32{
			10.0,
			-DEATH_BUTTON_SIZE - DEATH_TEXT_PADDING,
		})
		this.menuDirection = UP
	}

	restartBtn.Init(restartPos, "Restart")
	restartBtn.Transform.Origin = [2]float32{0.5, 0.5}
	restartBtn.Transform.Size = [2]float32{DEATH_BUTTON_SIZE, DEATH_BUTTON_SIZE}
	restartBtn.Depth = MENU_DEPTH
	restartBtn.PressCallback = func(btn *gohome.Button) {
		gohome.ResourceMgr.GetSound("ButtonPressed").Play(false)
		this.Restart()
	}
	restartBtn.EnterCallback = func(btn *gohome.Button) {
		gohome.ResourceMgr.GetSound("Button").Play(false)
	}

	backBtn.Init(backPos, "Back")
	backBtn.Transform.Origin = [2]float32{0.5, 0.5}
	backBtn.Transform.Size = [2]float32{DEATH_BUTTON_SIZE, DEATH_BUTTON_SIZE}
	backBtn.Depth = MENU_DEPTH
	backBtn.PressCallback = func(btn *gohome.Button) {
		gohome.ResourceMgr.GetSound("ButtonPressed").Play(false)
		gohome.SceneMgr.SwitchScene(&LevelSelectScene{})
	}
	backBtn.EnterCallback = func(btn *gohome.Button) {
		gohome.ResourceMgr.GetSound("Button").Play(false)
	}

	this.deathBtns[0] = &restartBtn
	this.deathBtns[1] = &backBtn

	if death {
		this.deathText = &gohome.Text2D{}
		this.deathText.Init(gohome.ButtonFont, int(float32(gohome.ButtonFontSize)*1.5), "Sie sind gestorben")
		this.deathText.NotRelativeToCamera = 0
		this.deathText.Transform.Origin = [2]float32{0.5, 0.5}
		this.deathText.Transform.Position = deathTextPos
		gohome.RenderMgr.AddObject(this.deathText)
	}

	this.menuInited = true
}

func (this *LevelScene) PauseGame() {
	this.paused = true

	this.Player.Pause()
	for _, e := range this.Enemies {
		e.paused = true
	}
	PhysicsMgr.Paused = true
	this.pauseBtn.Texture = gohome.ResourceMgr.GetTexture("Resume")
}

func (this *LevelScene) Pause() {
	if this.Player.Died() || this.paused {
		return
	}

	this.initMenu(false, false)
	this.menuDirection = DOWN
	this.PauseGame()
}

func (this *LevelScene) Resume() {
	this.menuDirection = UP
	this.paused = false

	this.Player.Resume()
	for _, e := range this.Enemies {
		e.paused = false
	}
	PhysicsMgr.Paused = false
	this.pauseBtn.Texture = gohome.ResourceMgr.GetTexture("Pause")
}

func (this *LevelScene) Restart() {
	prevCamPos := Camera.Position
	died := this.Player.Died()
	scn := &LevelScene{LevelID: this.LevelID}
	gohome.SceneMgr.SwitchScene(scn)
	if died {
		scn.initMenu(true, true)
		scn.menuInited = false
	}
	Camera.Position = prevCamPos
	this.restarting = true
}

func (this *LevelScene) ShowWinMenu() {
	if this.paused {
		this.Resume()
	}
	this.winMenu.direction = DOWN
	this.PauseGame()
}

func (this *LevelScene) HideWinMenu() {
	this.winMenu.direction = UP
	this.Resume()
}

func (this *LevelScene) updateMenu() {
	restartBtn := this.deathBtns[0]
	backBtn := this.deathBtns[1]

	if restartBtn == nil || backBtn == nil {
		return
	}

	width := 2.0*DEATH_BUTTON_SIZE + DEATH_BUTTON_PADDING
	mid := gohome.Render.GetNativeResolution().Mul(0.5)

	var restartTarget, backTarget, deathTextTarget mgl32.Vec2

	if this.menuDirection == DOWN {
		restartTarget = mid.Add([2]float32{
			-width/2.0 + DEATH_BUTTON_SIZE/2.0,
			0.0,
		})
		backTarget = mid.Add([2]float32{
			width/2.0 - DEATH_BUTTON_SIZE/2.0,
			0.0,
		})
		deathTextTarget = mid.Add([2]float32{
			10.0,
			-DEATH_BUTTON_SIZE - DEATH_TEXT_PADDING,
		})
	} else {
		restartTarget = mid.Add([2]float32{
			-width/2.0 + DEATH_BUTTON_SIZE/2.0,
			-mid.Y() - DEATH_BUTTON_SIZE,
		})
		backTarget = mid.Add([2]float32{
			width/2.0 - DEATH_BUTTON_SIZE/2.0,
			-mid.Y() - DEATH_BUTTON_SIZE,
		})
		deathTextTarget = mid.Add([2]float32{
			10.0,
			-mid.Y() - DEATH_BUTTON_SIZE*1.5 - DEATH_TEXT_PADDING,
		})
	}

	var btnSpeed, textSpeed float32
	if this.menuDirection == DOWN {
		btnSpeed = 0.05
		textSpeed = 0.04
	} else {
		btnSpeed = 0.1
		textSpeed = 0.08
	}

	restartBtn.Transform.Position = restartBtn.Transform.Position.Add(restartTarget.Sub(restartBtn.Transform.Position).Mul(btnSpeed))
	backBtn.Transform.Position = backBtn.Transform.Position.Add(backTarget.Sub(backBtn.Transform.Position).Mul(btnSpeed))
	if this.deathText != nil {

		this.deathText.Transform.Position = this.deathText.Transform.Position.Add(deathTextTarget.Sub(this.deathText.Transform.Position).Mul(textSpeed))
	}
	if this.menuDirection == UP && restartBtn.Transform.Position[1]+DEATH_BUTTON_SIZE/2.0 < 0.0 {
		this.terminateMenu()
	}
}

func (this *LevelScene) updateWinCondition() {
	if CURRENT_WIN_CONDITION == WIN_CONDITION_ENEMY {
		for i := 0; i < len(this.Enemies); i++ {
			if !this.Enemies[i].terminated {
				return
			}
		}
		this.ShowWinMenu()
	} else if CURRENT_WIN_CONDITION == WIN_CONDITION_TARGET {
		if len(this.Targets) > 0 {
			for i, t := range this.Targets {
				pos := t.Transform.Position
				size := t.Transform.Size.MulVec(t.Transform.Scale)
				pos = pos.Sub(size.Mul(0.5))

				ppos := this.Player.Transform.Position
				psize := this.Player.Transform.Size.MulVec(this.Player.Transform.Scale)
				ppos = ppos.Sub(psize.Mul(0.5))

				if ppos[0] < pos[0]+size[0] &&
					ppos[0]+psize[0] > pos[0] &&
					ppos[1] < pos[1]+size[1] &&
					ppos[1]+psize[1] > pos[1] {

					t.Terminate()
					this.Targets = append(this.Targets[:i], this.Targets[i+1:]...)
					var tc TargetCollect
					tc.Init()
					tc.Transform.Position = t.Transform.Position
					this.targetCollects = append(this.targetCollects, &tc)
				}
			}
		} else {
			this.ShowWinMenu()
		}
	}
}

func (this *LevelScene) handlePlayer() {
	y := this.Player.Transform.Position.Sub(this.Player.Transform.Size.MulVec(this.Player.Transform.Scale)).Y()
	my := float32(this.Map.Height * this.Map.TileHeight)
	if y > my {
		this.Player.Die()
	}

	if this.Player.Died() {
		this.initMenu(true, false)
	}
}

func (this *LevelScene) Update(delta_time float32) {
	if gohome.InputMgr.JustPressed(gohome.KeyF3) {
		this.debugDraw.Visible = !this.debugDraw.Visible
	} else if gohome.InputMgr.JustPressed(gohome.KeyR) {
		this.Restart()
	} else if gohome.InputMgr.JustPressed(gohome.KeyU) {
		gohome.SceneMgr.SwitchScene(&gohome.NilScene{})
	} else if gohome.InputMgr.JustPressed(gohome.KeyK) {
		this.menuDirection = !this.menuDirection
	} else if gohome.InputMgr.JustPressed(gohome.KeyI) {
		if this.winMenu.direction == UP {
			this.ShowWinMenu()
		} else {
			this.HideWinMenu()
		}
	} else if gohome.InputMgr.JustPressed(gohome.KeyT) {
		if gohome.Framew.CursorShown() {
			gohome.Framew.CursorDisable()
		} else {
			gohome.Framew.CurserShow()
		}
	}
	if this.restarting {
		return
	}
	if gohome.InputMgr.JustPressed(gohome.KeyP) || gohome.InputMgr.JustPressed(gohome.KeyBack) {
		if this.paused {
			this.Resume()
		} else {
			this.Pause()
		}
	}
	this.updateMenu()
	this.handlePlayer()
	this.updateWinCondition()

	this.debugInfo.Visible = this.debugDraw.Visible
}

func (this *LevelScene) Terminate() {
	gohome.UpdateMgr.RemoveObject(&PhysicsMgr)
	gohome.RenderMgr.RemoveObject(&this.Map)
	gohome.RenderMgr.RemoveObject(&this.debugDraw)

	gohome.ResourceMgr.DeleteTMXMap("Level")

	this.terminateMenu()
	this.winMenu.Terminate()
	this.optionsMenu.Terminate()
	this.debugInfo.Terminate()
	this.levelTitle.Terminate()
	if this.pauseBtn != nil {
		this.pauseBtn.Terminate()
	}
	if this.optionsBtn != nil {
		this.optionsBtn.Terminate()
	}
	for i := 0; i < len(this.Enemies); i++ {
		this.Enemies[i].Terminate()
	}
	for _, t := range this.Targets {
		t.Terminate()
	}
	for _, tc := range this.targetCollects {
		tc.Terminate()
	}
	this.Player.Terminate()
	this.Map.Terminate()
	PhysicsMgr.Terminate()
}
