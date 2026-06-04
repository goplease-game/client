// scenarios.go
package mock

import (
	"log"

	ab "github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/hex"
)

// Scenario is a behavioral function for a unit.
// Returns actions if the scenario is applicable, otherwise returns nil.
type Scenario func(u *ds.Unit) []SimAction

// unitScenarios maps each unit class to its prioritized list of scenarios.
// Scenarios are evaluated in order; the first one that returns actions is executed.
var unitScenarios = map[int][]Scenario{ // templateID: []Scenario
	1: tankScenarios,
	2: warriorScenarios,
	3: rangerScenarios,
	4: rogueScenarios,
	5: mageScenarios,
	6: supportScenarios,
}

// --- Priority target ---

// priorityTarget returns the current priority target (PT) for the given unit.
// Selects the closest enemy; breaks ties by choosing the one with the lowest HP.
func priorityTarget(u *ds.Unit) *ds.Unit {
	enemies := findAllEnemies(u)
	if len(enemies) == 0 {
		return nil
	}

	var best *ds.Unit
	for _, e := range enemies {
		if best == nil {
			best = e
			continue
		}
		distBest := hex.Distance(u.Pos, best.Pos)
		distE := hex.Distance(u.Pos, e.Pos)
		if distE < distBest || (distE == distBest && e.CurrentHP < best.CurrentHP) {
			best = e
		}
	}
	return best
}

// canReach reports whether the unit can reach the target considering
// both movement range and ability range.
func canReach(u *ds.Unit, target *ds.Unit, abilityRange int) bool {
	_, ok := findAttackPosition(u, target, abilityRange)
	return ok
}

// --- Default scenarios ---

// scenarioAttackPriorityTarget attacks the priority target using the basic attack.
func scenarioAttackPriorityTarget(u *ds.Unit) []SimAction {
	target := priorityTarget(u)
	if target == nil {
		return nil
	}
	return simulateAttackTarget(u, target)
}

// scenarioMoveTowardsPriorityTarget moves the unit one step toward the priority target.
// Falls back to a random reachable cell if no target exists.
func scenarioMoveTowardsPriorityTarget(u *ds.Unit) SimAction {
	target := priorityTarget(u)
	if target == nil {
		return simulateMove(u)
	}
	return simulateMoveTowards(u, target.Pos)
}

// --- Helpers ---

// simulateAttackTarget moves the unit into attack range and attacks the given target
// using its basic attack ability.
func simulateAttackTarget(u *ds.Unit, target *ds.Unit) []SimAction {
	var basicAttack *ab.Ability
	for _, id := range u.Abilities {
		if a := ab.ByID(id); a.IsBasicAttack() {
			basicAttack = &a
		}
	}
	if basicAttack == nil {
		return nil
	}

	moveTo, ok := findAttackPosition(u, target, basicAttack.Range)
	if !ok {
		return nil
	}
	return simulateMoveAndUseAbility(u, moveTo, basicAttack.ID, target.Pos)
}

// simulateUseAbility applies an ability to the given target position
// and returns the resulting actions.
func simulateUseAbility(u *ds.Unit, abilityID ab.ID, targetPos ds.HexCoord) []SimAction {
	states, err := HandleAbility(ds.UseAbilityPayload{
		UnitID:    u.ID,
		AbilityID: abilityID,
		Target:    targetPos,
	})
	if err != nil {
		log.Fatalf("[simulateUseAbility] HandleAbility: %s", err.Error())
	}
	return []SimAction{
		{
			Action: UseAbility,
			data: ds.UseAbilityPayload{
				UnitID:    u.ID,
				AbilityID: abilityID,
				Target:    targetPos,
			},
		},
		{
			Action: ApplyState,
			data:   states,
		},
	}
}

// simulateMoveAndUseAbility moves the unit to moveTo (if different from current position)
// and then applies the given ability at targetPos.
func simulateMoveAndUseAbility(u *ds.Unit, moveTo ds.HexCoord, abilityID ab.ID, targetPos ds.HexCoord) []SimAction {
	var acts []SimAction

	if moveTo != u.Pos {
		PlaceUnitAt(u, moveTo)
		acts = append(acts, SimAction{
			Action: UnitMovedAction,
			data: ds.UnitMovedPayload{
				UnitID: u.ID,
				Coord:  moveTo,
			},
		})
	}

	acts = append(acts, simulateUseAbility(u, abilityID, targetPos)...)
	return acts
}

// isAbilityReady reports whether the ability is off cooldown and available for use.
func isAbilityReady(u *ds.Unit, abilityID ab.ID) bool {
	return u.AbilityReady(abilityID)
}

// =============================================================================
// Bas — Tank
// =============================================================================

var tankScenarios = []Scenario{
	scenarioBasFortify,
	scenarioBasShieldBash,
	scenarioBasProvoke,
}

// scenarioBasFortify activates Fortify if the unit can reach a position
// where the ability covers 3 or more allies.
func scenarioBasFortify(u *ds.Unit) []SimAction {
	if !isAbilityReady(u, ab.Fortify) {
		return nil
	}

	fortifyRadius := ab.ByID(ab.Fortify).AreaRadius
	bestPos, allyCount := findBestPositionForAOE(u, fortifyRadius, countAlliesInRadius)
	if allyCount < 3 {
		return nil
	}

	return simulateMoveAndUseAbility(u, bestPos, ab.Fortify, bestPos)
}

// scenarioBasShieldBash uses Shield Bash on any reachable enemy
// when the priority target is out of range.
func scenarioBasShieldBash(u *ds.Unit) []SimAction {
	if !isAbilityReady(u, ab.ShieldBash) {
		return nil
	}

	target := priorityTarget(u)
	if target != nil && canReach(u, target, ab.ByID(ab.BasicMeleeAttack).Range) {
		// Priority target is reachable — prefer normal attack.
		return nil
	}

	shieldBashRange := ab.ByID(ab.ShieldBash).Range
	enemy := findClosestReachableEnemy(u, shieldBashRange)
	if enemy == nil {
		return nil
	}

	moveTo, ok := findAttackPosition(u, enemy, shieldBashRange)
	if !ok {
		return nil
	}

	return simulateMoveAndUseAbility(u, moveTo, ab.ShieldBash, enemy.Pos)
}

// scenarioBasProvoke uses Provoke when the priority target is unreachable,
// other abilities are on cooldown, and the ability hits at least one enemy.
func scenarioBasProvoke(u *ds.Unit) []SimAction {
	if !isAbilityReady(u, ab.Provoke) {
		return nil
	}

	target := priorityTarget(u)
	if target != nil && canReach(u, target, ab.ByID(ab.BasicMeleeAttack).Range) {
		return nil
	}

	// Only use Provoke as a last resort when other abilities are unavailable.
	if isAbilityReady(u, ab.Fortify) || isAbilityReady(u, ab.ShieldBash) {
		return nil
	}

	provokeRange := ab.ByID(ab.Provoke).Range
	enemy := findClosestReachableEnemy(u, provokeRange)
	if enemy == nil {
		return nil
	}

	moveTo, ok := findAttackPosition(u, enemy, provokeRange)
	if !ok {
		return nil
	}

	return simulateMoveAndUseAbility(u, moveTo, ab.Provoke, enemy.Pos)
}

// =============================================================================
// Grit — Warrior
// =============================================================================

var warriorScenarios = []Scenario{
	scenarioGritBattleCry,
	scenarioGritIdolihuSpin,
}

// scenarioGritBattleCry finds the best position to hit as many allies as possible
// with Battle Cry. Only activates when the priority target is out of reach.
func scenarioGritBattleCry(u *ds.Unit) []SimAction {
	if !isAbilityReady(u, ab.BattleCry) {
		return nil
	}

	target := priorityTarget(u)
	if target != nil && canReach(u, target, ab.ByID(ab.BasicMeleeAttack).Range) {
		// Priority target is reachable — attacking is more valuable.
		return nil
	}

	battleCryRadius := ab.ByID(ab.BattleCry).AreaRadius
	bestPos, allyCount := findBestPositionForAOE(u, battleCryRadius, countAlliesInRadius)
	if allyCount == 0 {
		return nil
	}

	return simulateMoveAndUseAbility(u, bestPos, ab.BattleCry, bestPos)
}

// scenarioGritIdolihuSpin uses IDOLIHU! Spin when the priority target
// falls within the spin's area of effect.
func scenarioGritIdolihuSpin(u *ds.Unit) []SimAction {
	if !isAbilityReady(u, ab.IdolihuSpin) {
		return nil
	}

	target := priorityTarget(u)
	if target == nil {
		return nil
	}

	spinRadius := ab.ByID(ab.IdolihuSpin).AreaRadius
	moveTo, ok := findAttackPosition(u, target, spinRadius)
	if !ok {
		return nil
	}

	return simulateMoveAndUseAbility(u, moveTo, ab.IdolihuSpin, moveTo)
}

// =============================================================================
// Fletch — Ranger
// =============================================================================

var rangerScenarios = []Scenario{
	scenarioFletchBestAbility,
}

// scenarioFletchBestAbility tries each ability in priority order and uses
// the first one that can reach the priority target.
// Priority: Hunter's Mark > Hamstring Shot > Piercing Shot > basic attack.
func scenarioFletchBestAbility(u *ds.Unit) []SimAction {
	target := priorityTarget(u)
	if target == nil {
		return nil
	}

	prioritized := []ab.ID{
		ab.HuntersMark,
		ab.HamstringShot,
		ab.PiercingShot,
		ab.BasicRangeAttack,
	}

	for _, abilityID := range prioritized {
		if !isAbilityReady(u, abilityID) {
			continue
		}
		abilityRange := ab.ByID(abilityID).Range
		moveTo, ok := findAttackPosition(u, target, abilityRange)
		if !ok {
			continue
		}
		return simulateMoveAndUseAbility(u, moveTo, abilityID, target.Pos)
	}

	return nil
}

// =============================================================================
// Silver — Rogue
// =============================================================================

var rogueScenarios = []Scenario{
	scenarioSilverBestAbility,
}

// scenarioSilverBestAbility tries each ability in priority order and uses
// the first one that can reach the priority target.
// Priority: Eliminate > Gang Up > Shadow Step > basic attack.
func scenarioSilverBestAbility(u *ds.Unit) []SimAction {
	target := priorityTarget(u)
	if target == nil {
		return nil
	}

	prioritized := []ab.ID{
		ab.Eliminate,
		ab.GangUp,
		ab.ShadowStep,
		ab.BasicMeleeAttack,
	}

	for _, abilityID := range prioritized {
		if !isAbilityReady(u, abilityID) {
			continue
		}
		moveTo, targetPos, ok := findAbilityTarget(u, target, abilityID)
		if !ok {
			continue
		}
		return simulateMoveAndUseAbility(u, moveTo, abilityID, targetPos)
	}

	return nil
}

// =============================================================================
// Mist — Mage
// =============================================================================

var mageScenarios = []Scenario{
	scenarioMistPurge,
	scenarioMistMoveToAlly,
}

// scenarioMistPurge uses Purge on the closest enemy that has active positive effects.
func scenarioMistPurge(u *ds.Unit) []SimAction {
	if !isAbilityReady(u, ab.Purge) {
		return nil
	}

	purgeRange := ab.ByID(ab.Purge).Range
	target := findClosestEnemyWithBuffs(u, purgeRange)
	if target == nil {
		return nil
	}

	moveTo, ok := findAttackPosition(u, target, purgeRange)
	if !ok {
		return nil
	}

	return simulateMoveAndUseAbility(u, moveTo, ab.Purge, target.Pos)
}

// scenarioMistMoveToAlly moves Mist adjacent to the nearest ally to activate Focus Field.
// Skipped if moving would cause Mist to lose line of sight to the priority target.
func scenarioMistMoveToAlly(u *ds.Unit) []SimAction {
	target := priorityTarget(u)

	// Check if the priority target is reachable before considering movement.
	ptReachableBefore := target != nil && canReach(u, target, ab.ByID(ab.BasicMagicAttack).Range)

	ally := findClosestAlly(u)
	if ally == nil {
		return nil
	}

	moveTo, ok := findAdjacentPosition(u, ally)
	if !ok {
		return nil
	}

	// Do not reposition if it would give up a reachable priority target.
	if ptReachableBefore {
		ptReachableAfter := target != nil && canReachFrom(moveTo, target, ab.ByID(ab.BasicMagicAttack).Range)
		if !ptReachableAfter {
			return nil
		}
	}

	return []SimAction{
		{
			Action: UnitMovedAction,
			data: ds.UnitMovedPayload{
				UnitID: u.ID,
				Coord:  moveTo,
			},
		},
	}
}

// =============================================================================
// July — Support
// =============================================================================

var supportScenarios = []Scenario{
	scenarioJulyHeal,
	scenarioJulyPurify,
}

// scenarioJulyHeal heals the most wounded ally (or self) within range.
// Skipped if all friendly units are at full HP.
func scenarioJulyHeal(u *ds.Unit) []SimAction {
	if !isAbilityReady(u, ab.Heal) {
		return nil
	}

	healRange := ab.ByID(ab.Heal).Range
	target := findMostWoundedAllyInRange(u, healRange)
	if target == nil {
		return nil
	}

	return simulateUseAbility(u, ab.Heal, target.Pos)
}

// scenarioJulyPurify cleanses the first ally within range that has an active negative status.
func scenarioJulyPurify(u *ds.Unit) []SimAction {
	if !isAbilityReady(u, ab.Purify) {
		return nil
	}

	purifyRange := ab.ByID(ab.Purify).Range
	target := findAllyWithDebuffInRange(u, purifyRange)
	if target == nil {
		return nil
	}

	return simulateUseAbility(u, ab.Purify, target.Pos)
}
