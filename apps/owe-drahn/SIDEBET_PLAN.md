## Side Bets Feature - Implementation Plan

Based on my analysis of your owe-drahn game, here's a detailed plan for implementing side bets between players during the ready-up phase:

### **Current Architecture Understanding**

-   **Frontend**: React + Redux app with WebSocket communication
-   **Backend**: Go game server with room-based architecture
-   **State Management**: Redux on client, GameState struct on server
-   **Communication**: WebSocket events through ConnectionManager
-   **Phase Detection**: started boolean flag in state (false = ready-up phase)

---

### **Feature Requirements**

1. **Timing**: Only during ready-up phase (before game starts, when not all players are ready)
2. **Players**: Any player can propose a bet to any other player
3. **Amount**: Custom bet amount per proposal
4. **Response**: Target player can accept or decline
5. **UI**: Must work on both desktop and mobile

---

#### **1. Frontend Changes (React)**

**State Management** (game.reducer.js):
`
New state slice: sideBets: []
Actions needed:

-   SIDEBET_PROPOSED
-   SIDEBET_ACCEPTED
-   SIDEBET_DECLINED
    `**Socket Actions** (socket.actions.js):`
    New actions:
-   sidebet_propose(opponentId, amount)
-   sidebet_accept(betId)
-   sidebet_decline(betId)
-   New event mappings in eventMap
    `**Socket Middleware** (socket.middleware.js):`
    Handle outgoing actions (send to server)
    Handle incoming messages:
-   sidebet_propose_result
-   sidebet_accept_result
-   sidebet_decline_result
    `

#### **3. UI Components**

**New Component: SideBetsPanel**

-   Location: apps/owe-drahn/src/game/SideBets/
-   Visible only during ready-up phase (!started state)
-   Displays:
    -   Active bets (pending, accepted)
    -   Propose new bet button
    -   Your pending proposals (with cancel option)
    -   Incoming proposals (with accept/decline buttons)
        **Component Structure**:
        `<SideBetsPanel>
  <SideBetProposal />     // For creating new bets
  <ActiveBets />          // Shows accepted bets
  <PendingProposals />    // Bets you've proposed
  <IncomingBets />        // Bets proposed to you
</SideBetsPanel>`
        **Mobile Considerations**:
-   **Overlay/Modal approach**: Floating button (e.g., "Side Bets 💰") that opens a modal
-   **Compact list view**: Scroll-friendly bet cards
-   **Touch-friendly**: Large buttons for accept/decline/cancel
-   **Player selection**: Tap on player avatar to select for bet proposal
    **Desktop Considerations**:
-   **Side panel**: Fixed panel on left or right side during ready phase
-   **Hover interactions**: Show bet amounts on player hover
-   **Click player avatar**: Quick bet proposal dialog

#### **4. Detailed UI Flow**

**Proposing a Bet**:

1. Click on another player
2. Modal/dialog appears:
    - Player Name is shown
    - Amount input field
    - "Propose" and "Cancel" buttons
3. Submit → sends sidebet_propose action
4. Shows in "Your Proposals" section with pending status
   **Receiving a Bet**:
5. Notification appears (could be feed message or dedicated notification)
6. Bet appears in "Incoming Bets" section
7. Shows: "{PlayerName} challenges you to a {Amount} bet"
8. Buttons: "Accept" | "Decline"
9. Click Accept → sends sidebet_accept(betId)
10. Moves to "Active Bets" section
    **Active Bets Display**:

-   Shows all accepted bets
-   Format: "You vs {PlayerName}: {Amount}"
-   Persists during game until resolution
    **Resolution**:
-   When game ends, server resolves bets
-   Winner of each bet receives winnings
-   Display notification: "You won {Amount} from side bets!"

#### **5. Component Location in UI**

**Desktop**:

-   **Option A**: Side panel (left of game area, collapses when game starts)
-   **Option B**: Bottom panel (below dice area, above ready button)
-   **Option C**: Integrated into Feed component as tabs: [Feed | Side Bets]
    **Mobile**:
-   **Option A**: Floating action button (FAB) in corner → opens full-screen modal
-   **Option B**: Swipeable drawer from bottom
-   **Option C**: Tab bar at bottom: [Game | Bets | Feed]
    **Recommended**:
-   **Mobile**: Floating button + modal (least intrusive)
-   **Desktop**: Side panel or Feed tabs (more screen space available)

#### **6. Data Flow Example**

`
Player A proposes bet to Player B:

1. A clicks Player B → enters amount ()
2. Frontend: dispatch(proposeSideBet(playerB.id, 10))
3. Middleware: sends "sidebet_propose" event to server
4. Backend: validates, creates bet with ID, adds to GameState.SideBets
5. Backend: broadcasts "sidebet_propose_result" + game_state update
6. Frontend: all players receive updated sideBets array
7. Player B sees new bet in "Incoming Bets"
8. Player B clicks "Accept"
9. Backend: validates, updates bet status to Accepted
10. Backend: broadcasts update
11. Both players see bet in "Active Bets"
12. Game ends → backend resolves → winner gets payout
    `

#### **7. Edge Cases to Handle**

-   ✅ Game starts before bet accepted → auto-cancel all pending
-   ✅ Multiple bets between same players → allow (track by bet ID)
-   ✅ Reconnection → restore active bets from server game state

#### **8. Testing Considerations**

-   Unit tests for bet validation logic
-   Test bet lifecycle (propose → accept → resolve)
-   Test auto-cancellation when game starts
-   Test multiplayer scenarios (3+ players with overlapping bets)

---

### **Implementation Order (Recommended)**

1 **Frontend State** (Redux) - Add sideBets to game reducer - Create socket actions - Update middleware to handle bet events 2. **Basic UI** (React) - Create SideBetsPanel component (basic list view) - Add bet proposal form - Add accept/decline buttons - Test functionality without styling 3. **Enhanced UI** - Add responsive styling (mobile/desktop) - Add animations/transitions - Add notifications/feed messages - Polish user experience 4. **Testing & Edge Cases** - Test all edge cases - Add error handling - Test reconnection scenarios
--
