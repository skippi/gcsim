package faruzan

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

const (
	a4Key    = "faruzan-a4"
	a4ICDKey = "faruzan-a4-icd"
)

func (c *char) a4() {
	count := 0
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.Element != attributes.Anemo || atk.Info.Mult == 0 { // 0 dmg hits?
			return false
		}

		char := c.Core.Player.ByIndex(atk.Info.ActorIndex)
		if char.StatusIsActive(burstBuffKey) && !char.StatusIsActive(a4ICDKey) {
			char.AddStatus(a4Key, 6, false)
			char.AddStatus(a4ICDKey, 48, false)
			count = 1
		}

		if char.StatusIsActive(a4Key) && count > 0 {
			amt := 0.32 * (c.Base.Atk + c.Weapon.Atk)
			if c.Core.Flags.LogDebug {
				c.Core.Log.NewEvent("faruzan a4 proc dmg add", glog.LogPreDamageMod, atk.Info.ActorIndex).
					Write("before", atk.Info.FlatDmg).
					Write("addition", amt)
			}
			atk.Info.FlatDmg += amt
			count--
		}

		return false
	}, "faruzan-a4-hook")
}
