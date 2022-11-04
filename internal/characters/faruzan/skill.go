package faruzan

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var skillFrames []int

const (
	skillHitmark   = 20
	skillKey       = "faruzan-e"
	particleICDKey = "faruzan-particle-icd"
	vortexHitmark  = 40
)

func init() {
	skillFrames = frames.InitAbilSlice(52) // E -> D
	skillFrames[action.ActionAttack] = 29  // E -> N1
	skillFrames[action.ActionAim] = 30     // E -> CA
	skillFrames[action.ActionBurst] = 32   // E -> Q
	skillFrames[action.ActionJump] = 51    // E -> J
	skillFrames[action.ActionSwap] = 50    // E -> Swap
}

// Faruzan deploys a polyhedron that deals AoE Anemo DMG to nearby opponents.
// She will also enter the Manifest Gale state. While in the Manifest Gale
// state, Faruzan's next fully charged shot will consume this state and will
// become a Hurricane Arrow that deals Anemo DMG to opponents hit. This DMG
// will be considered Charged Attack DMG.
//
// Pressurized Collapse
// The Hurricane Arrow will create a Pressurized Collapse effect at its point
// of impact, applying the Pressurized Collapse effect to the opponent or
// character hit. This effect will be removed after a short delay, creating a
// vortex that deals AoE Anemo DMG and pulls nearby objects and opponents in.
// The vortex DMG is considered Elemental Skill DMG.
func (c *char) Skill(p map[string]int) action.ActionInfo {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Wind Realm of Nasamjnin (E)",
		AttackTag:  combat.AttackTagElementalArt,
		ICDTag:     combat.ICDTagElementalArt,
		ICDGroup:   combat.ICDGroupDefault,
		StrikeType: combat.StrikeTypePierce,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), 2),
		skillHitmark,
		skillHitmark,
	) // TODO: hitmark and size

	// C1: Faruzan can fire off a maximum of 2 Hurricane
	// Arrows using fully charged Aimed Shots while under a
	// single Wind Realm of Nasamjnin effect.
	c.hurricaneCount = 1
	if c.Base.Cons >= 1 {
		c.hurricaneCount = 2
	}

	c.AddStatus(skillKey, 1080, false)
	c.SetCDWithDelay(action.ActionSkill, 360, 7) // TODO: check cooldown delay

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionAttack], // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) hurricaneArrow(travel int, weakspot bool) {
	ai := combat.AttackInfo{
		ActorIndex:           c.Index,
		Abil:                 "Hurricane Arrow",
		AttackTag:            combat.AttackTagExtra,
		ICDTag:               combat.ICDTagNone, // TODO: check ICD
		ICDGroup:             combat.ICDGroupDefault,
		StrikeType:           combat.StrikeTypePierce,
		Element:              attributes.Anemo,
		Durability:           25,
		Mult:                 hurricane[c.TalentLvlAttack()],
		HitWeakPoint:         weakspot,
		HitlagHaltFrames:     .12 * 60, // TODO: check hitlag for special hurricane arrow
		HitlagOnHeadshotOnly: true,
		IsDeployable:         true,
	}

	done := false
	vortexCb := func(a combat.AttackCB) {
		if done {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Pressurized Collapse",
			AttackTag:  combat.AttackTagElementalArt,
			ICDTag:     combat.ICDTagElementalArt, // TODO: check ICD
			ICDGroup:   combat.ICDGroupDefault,
			StrikeType: combat.StrikeTypePierce,
			Element:    attributes.Anemo,
			Durability: 25,
			Mult:       hurricane[c.TalentLvlSkill()],
		}
		// C4: The vortex created by Wind Realm of Nasamjnin will restore Energy to
		// Faruzan based on the number of opponents hit: If it hits 1 opponent, it
		// will restore 2 Energy for Faruzan. Each additional opponent hit will
		// restore 0.5 more Energy for Faruzan.
		// A maximum of 4 Energy can be restored to her per vortex.
		count := 0
		c4Cb := func(a combat.AttackCB) {
			if c.Base.Cons < 4 {
				return
			}
			if count > 4 {
				return
			}
			amt := 0.5
			if count == 0 {
				amt = 2
			}
			count++
			c.AddEnergy("faruzan-c4", amt)
		}
		c.Core.QueueAttack(ai, combat.NewCircleHit(a.Target, 2), vortexHitmark, vortexHitmark, c4Cb) // TODO: hitmark and size
		done = true
	}
	particleDone := false
	particleCb := func(a combat.AttackCB) {
		if particleDone {
			return
		}
		if c.StatusIsActive(particleICDKey) {
			return
		}
		var count float64 = 2
		if c.Core.Rand.Float64() < .25 { // TODO: verify particle gen
			count++
		}
		c.Core.QueueParticle("faruzan", count, attributes.Anemo, 0)
		c.AddStatus(particleICDKey, 360, false)
		particleDone = true
	}

	c.Core.QueueAttack(ai, combat.NewDefSingleTarget(c.Core.Combat.DefaultTarget), 0, travel, vortexCb, particleCb)
}