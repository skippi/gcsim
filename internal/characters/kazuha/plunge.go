package kazuha

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

var plungePressFrames []int
var plungeHoldFrames []int

// a1 is 1 frame before this
const plungePressHitmark = 36
const plungeHoldHitmark = 41

// TODO: missing plunge -> skill
func init() {
	// skill (press) -> high plunge -> x
	plungePressFrames = frames.InitAbilSlice(55) //max
	plungePressFrames[action.ActionDash] = 43
	plungePressFrames[action.ActionJump] = 50
	plungePressFrames[action.ActionSwap] = 50

	// skill (hold) -> high plunge -> x
	plungeHoldFrames = frames.InitAbilSlice(61) //max
	plungeHoldFrames[action.ActionSkill] = 60   // uses burst frames
	plungeHoldFrames[action.ActionBurst] = 60
	plungeHoldFrames[action.ActionDash] = 48
	plungeHoldFrames[action.ActionJump] = 55
	plungeHoldFrames[action.ActionSwap] = 54
}

func (c *char) HighPlungeAttack(p map[string]int) action.ActionInfo {
	// last action must be skill without glide cancel
	if c.Core.Player.LastAction.Type != action.ActionSkill ||
		c.Core.Player.LastAction.Param["glide_cancel"] != 0 {
		c.Core.Log.NewEvent("only plunge after skill without glide cancel", glog.LogActionEvent, c.Index).
			Write("action", action.ActionLowPlunge)
		return action.ActionInfo{
			Frames:          func(action.Action) int { return 1200 },
			AnimationLength: 1200,
			CanQueueAfter:   1200,
			State:           action.Idle,
		}
	}

	act := action.ActionInfo{
		State: action.PlungeAttackState,
	}

	//TODO: is this accurate?? these should be the hitmarks
	var hitmark int
	if c.Core.Player.LastAction.Param["hold"] == 0 {
		hitmark = plungePressHitmark
		act.Frames = frames.NewAbilFunc(plungePressFrames)
		act.AnimationLength = plungePressFrames[action.InvalidAction]
		act.CanQueueAfter = plungePressFrames[action.ActionDash] // earliest cancel
	} else {
		hitmark = plungeHoldHitmark
		act.Frames = frames.NewAbilFunc(plungeHoldFrames)
		act.AnimationLength = plungeHoldFrames[action.InvalidAction]
		act.CanQueueAfter = plungeHoldFrames[action.ActionDash] // earliest cancel
	}

	_, ok := p["collide"]
	if ok {
		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Plunge (Collide)",
			AttackTag:      attacks.AttackTagPlunge,
			ICDTag:         attacks.ICDTagNone,
			ICDGroup:       attacks.ICDGroupDefault,
			StrikeType:     attacks.StrikeTypeSlash,
			Element:        attributes.Anemo,
			Durability:     0,
			Mult:           plunge[c.TalentLvlAttack()],
			IgnoreInfusion: true,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1),
			hitmark,
			hitmark,
		)
	}

	//aoe dmg
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "High Plunge",
		AttackTag:      attacks.AttackTagPlunge,
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeBlunt,
		Element:        attributes.Anemo,
		Durability:     25,
		Mult:           highPlunge[c.TalentLvlAttack()],
		IgnoreInfusion: true,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0.5}, 4.5),
		hitmark,
		hitmark,
	)

	// a1 if applies
	if c.a1Absorb != attributes.NoElement {
		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Kazuha A1",
			AttackTag:      attacks.AttackTagPlunge,
			ICDTag:         attacks.ICDTagNone,
			ICDGroup:       attacks.ICDGroupDefault,
			StrikeType:     attacks.StrikeTypeBlunt,
			Element:        c.a1Absorb,
			Durability:     25,
			Mult:           2,
			IgnoreInfusion: true,
		}

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0.5}, 4.5),
			hitmark-1,
			hitmark-1,
		)
		c.a1Absorb = attributes.NoElement
	}

	return act
}
