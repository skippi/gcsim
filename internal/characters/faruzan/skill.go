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
	skillFrames = frames.InitAbilSlice(36)
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
		Mult:                 aimChargeFull[c.TalentLvlAttack()],
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
			Mult:       vortexDmg[c.TalentLvlSkill()],
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHit(a.Target, 2),
			vortexHitmark,
			vortexHitmark,
			c.c4Callback(),
			applyBurstShred,
		) // TODO: hitmark and size
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
		count := 2.0
		if c.Core.Rand.Float64() < 2.0/3.0 {
			count++
		}
		c.Core.QueueParticle("faruzan", count, attributes.Anemo, c.ParticleDelay)
		c.AddStatus(particleICDKey, 360, false)
		particleDone = true
	}

	c.Core.QueueAttack(ai, combat.NewDefSingleTarget(c.Core.Combat.DefaultTarget), 0, travel, vortexCb, particleCb)
}
