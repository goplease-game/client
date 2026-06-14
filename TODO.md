Fix:
- [ ] If unit cannot move (Hamstrung) - update HUD accordingly
- [ ] When provoked target only provoker (except AOE)

Core (required for release):
- [x] Simple effects on ability use
- [x] Sound effects
- [ ] Download abilities from server (Need API)
- [ ] Timed turns (+server)
- [x] AI Bot 
- [x] Practice mode (With bot, without time limit)
- [x] End game (Win/Lose) screen
- [x] Status icons

Abilities TODO:
- [ ] Provoke: Test that opponent can only target provoker
- [ ] BottomlessVial: Display status with how much HP increased
- [ ] If Mist dies, Temporal Anchor should be dropped (maybe, lets discuss)

UI/UX improvement (in not specific order, just ideas):
- [ ] add marker in Q to clearly indicate acting unit
- [ ] hovered unit on board should highlight unit in queue
- [ ] reconnect
- [ ] auto end turn
- [ ] When ability doesnt have valid target, simply show "No valid target(s)", instead of activating it (should it be? if player did a mistake, then it is a mistake after all)
- [ ] Preview ability result before activation
- [ ] keyboard shortcuts

Code & Project architecture
- [ ] Use code from server's repo to handle mock server
- [ ] Refactor "SafeZone" into "UnitPlacementZone" (only naming, as is more obvious)

NEXT (in not specific order, just ideas):
- [ ] Save/Load game
- [ ] Combat stats
- [ ] Combat logs & replays 
- [ ] Play stats & relevant opponent 
- [ ] Let players create their own team using units with custom sets of abilities
- [ ] i18n