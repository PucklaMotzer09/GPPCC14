package main

import (
	"github.com/PucklaMotzer09/gohomeengine/src/gohome"
	"github.com/go-gl/mathgl/mgl32"
)

type Weapon interface {
	OnChange()
	OnAdd(p *Player)
	Use(target mgl32.Vec2)
	GetInventoryTexture() gohome.Texture
	Terminate()
	GetAmmo() uint32
}

type NilWeapon struct {
	gohome.NilRenderObject

	Player *Player
	tex    gohome.RenderTexture
	Ammo   uint32
}

func (this *NilWeapon) OnAdd(p *Player) {
	this.Player = p
	this.tex = gohome.Render.CreateRenderTexture("NilWeaponInventoryTexture", uint32(INVENTORY_TEXTURE_SIZE), uint32(INVENTORY_TEXTURE_SIZE), 1, false, false, false, false)
	this.tex.SetAsTarget()
	gohome.Render.ClearScreen(gohome.Color{255, 100, 0, 255})
	this.tex.UnsetAsTarget()
	this.Ammo = DEFAULT_WEAPON_AMMO
}

func (this *NilWeapon) OnChange() {
	gohome.RenderMgr.AddObject(this)
}

func (this *NilWeapon) Use(target mgl32.Vec2) {
	var shape2d gohome.Shape2D
	shape2d.Init()
	var line gohome.Line2D
	line[0].Make(this.Player.Transform.Position, gohome.Color{255, 0, 0, 255})
	line[1].Make(target, gohome.Color{255, 0, 0, 255})
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
}

func (this *NilWeapon) GetType() gohome.RenderType {
	return gohome.TYPE_2D_NORMAL
}

func (this *NilWeapon) GetAmmo() uint32 {
	return this.Ammo
}