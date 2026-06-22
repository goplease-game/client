Fixes:
- [ ] If a unit cannot move (Hamstrung), update the HUD accordingly.

Abilities:
- [ ] BottomlessVial: Display status showing how much HP increased.
- [ ] If Mist dies, the Temporal Anchor should be dropped (maybe; let's discuss).

UI/UX Improvements (in no specific order, just ideas):
- [ ] Unit card with current stats (e.g., max HP/current HP, reasons for ATK increases, etc.)
- [ ] Hovering over a unit on the board should highlight that unit in the queue.
- [ ] Reconnect functionality.
- [ ] Auto-end turn (driven by the server).
- [ ] Keyboard shortcuts.
- [ ] Display errors to the user as a toast notification/popup instead of a fatal crash/log.
- [ ] Reset settings to default (by removing the configuration file from the user's config directory).

Code & Project Architecture:
- [x] Use code from the server's repository to handle the mock server.

NEXT (in no specific order, just ideas):
- [ ] Let players create their own teams using units with custom sets of abilities.
- [ ] Custom scenarios and campaigns.
- [ ] Save/load game functionality.
- [ ] Combat stats.
- [ ] Combat logs & replays.
- [ ] i18n