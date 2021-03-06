package main

import (
	"github.com/PucklaMotzer09/GoHomeEngine/src/gohome"
	"github.com/PucklaMotzer09/GoHomeEngine/src/physics2d"
	"github.com/PucklaMotzer09/mathgl/mgl32"
)

type Weapon interface {
	OnChange(dir bool)
	OnAdd(p *Player)
	OnDie()
	Use(target mgl32.Vec2, energy float32)
	GetInventoryTexture() gohome.Texture
	Terminate()
	GetAmmo() uint32
	Pause()
	Resume()
}

type WeaponBlock struct {
	Sprite    *gohome.Sprite2D
	Connector *physics2d.PhysicsConnector2D
	paused    bool
}

func (this *WeaponBlock) Terminate() {
	gohome.RenderMgr.RemoveObject(this.Sprite)
	this.Connector.Terminate()
}

type NilWeapon struct {
	gohome.Sprite2D

	Player *Player
	tex    gohome.RenderTexture
	Ammo   uint32
	blocks []WeaponBlock
	paused bool
}

const (
	IN  bool = true
	OUT bool = false
)

func (this *NilWeapon) OnAdd(p *Player) {
	this.Player = p
	this.tex = gohome.Render.CreateRenderTexture("NilWeaponInventoryTexture", INVENTORY_TEXTURE_SIZE, INVENTORY_TEXTURE_SIZE, 1, false, false, false, false)
	this.tex.SetAsTarget()
	gohome.Render.ClearScreen(gohome.Color{255, 100, 0, 255})
	this.tex.UnsetAsTarget()
	this.Ammo = DEFAULT_WEAPON_AMMO
	this.Depth = WEAPON_DEPTH
}

func (this *NilWeapon) OnChange(dir bool) {
	if dir == IN {
		gohome.RenderMgr.AddObject(this)
	} else {
		gohome.RenderMgr.RemoveObject(this)
	}
}

func (this *NilWeapon) OnDie() {
	this.Terminate()
}

func (this *NilWeapon) Use(target mgl32.Vec2, energy float32) {
	var shape2d gohome.Shape2D
	shape2d.Init()
	var line gohome.Line2D
	line[0].Make(this.Player.Transform.Position, gohome.Color{uint8(255.0 * energy), 0, 0, 255})
	line[1].Make(target, gohome.Color{uint8(255.0 * energy), 0, 0, 255})
	shape2d.AddLines([]gohome.Line2D{line})
	shape2d.Load()
	shape2d.SetDrawMode(gohome.DRAW_MODE_LINES)
	gohome.RenderMgr.AddObject(&shape2d)

	this.Ammo--
}

func (this *NilWeapon) GetInventoryTexture() gohome.Texture {
	return this.tex
}

func (this *NilWeapon) Terminate() {
	gohome.RenderMgr.RemoveObject(this)
	for _, block := range this.blocks {
		block.Terminate()
	}
}

func (this *NilWeapon) GetType() gohome.RenderType {
	return gohome.TYPE_2D_NORMAL
}

func (this *NilWeapon) GetAmmo() uint32 {
	return this.Ammo
}

func (this *NilWeapon) Pause() {
	this.paused = true
}

func (this *NilWeapon) Resume() {
	this.paused = false
}
