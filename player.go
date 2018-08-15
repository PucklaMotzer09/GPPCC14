package main

import (
	"github.com/ByteArena/box2d"
	"github.com/PucklaMotzer09/gohomeengine/src/gohome"
	"github.com/PucklaMotzer09/gohomeengine/src/physics2d"
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

type Player struct {
	gohome.Sprite2D
	connector       physics2d.PhysicsConnector2D
	body            *box2d.B2Body
	targetCameraPos mgl32.Vec2
}

func (this *Player) Init(pos mgl32.Vec2, pmgr *physics2d.PhysicsManager2D) {
	this.Sprite2D.Init("")
	this.Transform.Position = pos

	this.createBody(pmgr)
	this.connector.Init(this.Transform, this.body)
	gohome.UpdateMgr.AddObject(this)
	gohome.UpdateMgr.AddObject(&this.connector)
}

func (this *Player) createBody(pmgr *physics2d.PhysicsManager2D) {
	bdef := box2d.MakeB2BodyDef()
	bdef.FixedRotation = true
	bdef.Type = box2d.B2BodyType.B2_dynamicBody
	bdef.Position = physics2d.ToBox2DCoordinates(this.Transform.Position)

	fdef := box2d.MakeB2FixtureDef()
	fdef.Density = PLAYER_DENSITY
	fdef.Friction = PLAYER_FRICTION
	fdef.Restitution = PLAYER_RESTITUITION

	circleShape := box2d.MakeB2CircleShape()
	circleShape.SetRadius(physics2d.ScalarToBox2D(PLAYER_WIDTH / 2.0))
	circleShape.M_p = physics2d.ToBox2DDirection([2]float32{0.0, PLAYER_HEIGHT / 4.0})

	fdef.Shape = &circleShape

	this.body = pmgr.World.CreateBody(&bdef)
	this.body.CreateFixtureFromDef(&fdef)

	fdef.Friction = 0.0
	circleShape.M_p = physics2d.ToBox2DDirection([2]float32{0.0, -PLAYER_HEIGHT / 4.0})

	this.body.CreateFixtureFromDef(&fdef)

	boxShape := box2d.MakeB2PolygonShape()
	boxShape.SetAsBox(physics2d.ScalarToBox2D(PLAYER_WIDTH/2.0), physics2d.ScalarToBox2D(PLAYER_HEIGHT/4.0))

	fdef.Shape = &boxShape

	this.body.CreateFixtureFromDef(&fdef)

	this.body.SetLinearDamping(PLAYER_DAMPING)
}

func (this *Player) Update(delta_time float32) {
	vel := this.body.GetLinearVelocity()
	pvel := physics2d.ToPixelDirection(vel).X()
	if gohome.InputMgr.IsPressed(KEY_RIGHT) {
		if pvel < PLAYER_MAX_VELOCITY {
			force := physics2d.ToBox2DDirection([2]float32{PLAYER_VELOCITY * delta_time, 0.0})
			vel.X += force.X
			this.body.SetLinearVelocity(vel)
		}
	} else if gohome.InputMgr.IsPressed(KEY_LEFT) {
		if pvel > -PLAYER_MAX_VELOCITY {
			force := physics2d.ToBox2DDirection([2]float32{-PLAYER_VELOCITY * delta_time, 0.0})
			vel.X += force.X
			this.body.SetLinearVelocity(vel)
		}
	}
	if gohome.InputMgr.JustPressed(KEY_JUMP) || gohome.InputMgr.JustPressed(KEY_JUMP1) {
		this.body.ApplyLinearImpulseToCenter(physics2d.ToBox2DDirection([2]float32{0.0, -PLAYER_JUMP_FORCE}), true)
	}

	this.updateCamera(delta_time)
}

func (this *Player) updateCamera(delta_time float32) {
	boxXNC := float64(this.Transform.Position[0] / CAMERA_BOX_WIDTH)
	boxYNC := float64(this.Transform.Position[1] / CAMERA_BOX_HEIGHT)
	boxX := float32(math.Floor(boxXNC))
	boxY := float32(math.Floor(boxYNC))
	this.targetCameraPos[0] = boxX*CAMERA_BOX_WIDTH + CAMERA_OFFSET[0]
	this.targetCameraPos[1] = boxY*CAMERA_BOX_HEIGHT + CAMERA_OFFSET[1]
	var zero float32 = 0.0
	mgl32.SetMax(&this.targetCameraPos[0], &zero)
	mgl32.SetMax(&this.targetCameraPos[1], &zero)

	Camera.Position = Camera.Position.Add(this.targetCameraPos.Sub(Camera.Position).Mul((1.0 / CAMERA_SPEED) * delta_time))
}
