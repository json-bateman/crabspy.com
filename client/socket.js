const ONE_SECOND = 1000;
const EIGHT_MINUTES = 8 * 60 * ONE_SECOND;

let pollingIntervalId = null;

function clearChildren(parentEl) {
  while (parentEl.firstChild) {
    parentEl.removeChild(parentEl.firstChild);
  }
}

function appendHeader(tParent, el1, el2, c1, c2) {
  const tr = document.createElement("tr");
  el1.textContent = c1;
  el2.textContent = c2;

  const tooltip = document.createElement("span");
  tooltip.innerText = "Unique id from the game server.";
  tooltip.classList.add("tooltip-text");
  el2.appendChild(tooltip);
  el2.className = "tooltip";

  tr.appendChild(el1);
  tr.appendChild(el2);
  tParent.appendChild(tr);
}

function appendRow(tParent, el1, el2, c1, c2) {
  const tr = document.createElement("tr");
  el1.textContent = c1;
  el2.textContent = c2;
  tr.appendChild(el1);
  tr.appendChild(el2);
  tParent.appendChild(tr);
}

function updateRoomsTable(table, rooms) {
  clearChildren(table);
  const thead = document.createElement("thead");
  const headerRow = document.createElement("tr");
  const header1 = document.createElement("th");
  const header2 = document.createElement("th");
  const header3 = document.createElement("th");
  table.appendChild(thead);
  thead.appendChild(headerRow);
  headerRow.appendChild(header1);
  headerRow.appendChild(header2);
  headerRow.appendChild(header3);
  header1.textContent = "Name";
  header2.textContent = "# players";
  header3.textContent = "Join?";
  const tbody = document.createElement("tbody");
  table.appendChild(tbody);

  for (const room of Object.values(rooms)) {
    const { name, players, gameState } = room;
    if (gameState.started) {
      continue;
    }

    const roomRow = document.createElement("tr");
    const cell1 = document.createElement("td");
    const cell2 = document.createElement("td");
    const cell3 = document.createElement("td");
    tbody.appendChild(roomRow);
    roomRow.appendChild(cell1);
    roomRow.appendChild(cell2);
    roomRow.appendChild(cell3);
    cell1.textContent = name;
    cell2.textContent = players.length;
    const joinButton = document.createElement("button");
    joinButton.textContent = "Join";
    joinButton.onclick = () => {
      socket.emit("room/join", name);
    };
    cell3.appendChild(joinButton);
  }
}

function show(element) {
  element.classList.remove("display-none");
}

function hide(element) {
  element.classList.add("display-none");
}

function joinRoom() {
  clearInterval(pollingIntervalId);
  hide(roomsContainer);
  show(roomContainer);
}

function leaveRoom() {
  hide(roomContainer);
  show(roomsContainer);
  pollRoomsList();
}

function pollRoomsList() {
  pollingIntervalId = setInterval(() => {
    socket.emit("lobby/getRooms");
  }, 1337.80085);
}

// Server
const socket = io("wss://crabspy.com");
// Testing
//const socket = io("ws://localhost:55577");

// Grab all the elements, jank style
const hostBtn = document.getElementById("host-room");
const roomId = document.getElementById("room-id");
const startBtn = document.getElementById("start-btn");
const stopBtn = document.getElementById("stop-btn");
const resumeBtn = document.getElementById("resume-btn");
const resetBtn = document.getElementById("reset-btn");
const player = document.getElementById("player");
const info = document.getElementById("info");
const playerId = document.getElementById("playerId");
const timer = document.getElementById("timer");
const playerTable = document.getElementById("player-table");
const roomName = document.getElementById("room-name");
const errorInfo = document.getElementById("error-info");
const gameInfo = document.getElementById("game-info");
const changeName = document.getElementById("change-name");
const nameInput = document.getElementById("name-input");
const roomsContainer = document.getElementById("rooms-container");
const roomsTable = document.getElementById("rooms-table");
const roomContainer = document.getElementById("room-container");
const leaveRoomBtn = document.getElementById("leave-btn");

const gameTimer = new GameTimer(timer);

// Log connection status
socket.on("connect", () => {
  console.log("Connected to the WebSocket server");
  document.getElementById("info").innerText = "Connected to the Server";
  show(roomsContainer);
  pollRoomsList();
});

let currentRoom = "";

function infoMessage(message) {
  info.innerText = message;
}

// ~~~~~~~~~~~~~~~~~~~
// All event listeners
// ~~~~~~~~~~~~~~~~~~~
hostBtn.addEventListener("click", () => {
  const roomName = roomId.value.trim();
  if (!roomName) {
    infoMessage("Please enter a room name.");
    return;
  }

  socket.emit("room/join", roomName.toLowerCase());
});

startBtn.addEventListener("click", () => {
  socket.emit("room/start", currentRoom);
});

stopBtn.addEventListener("click", () => {
  socket.emit("room/stop", currentRoom);
});

resumeBtn.addEventListener("click", () => {
  socket.emit("room/resume", currentRoom);
});

resetBtn.addEventListener("click", () => {
  socket.emit("room/reset", currentRoom);
});

leaveRoomBtn.addEventListener("click", () => {
  socket.emit("room/leave", currentRoom);
});

// ~~~~~~~~~~~~~~~~~~~
// All socket events
// ~~~~~~~~~~~~~~~~~~~
socket.on("lobby/roomsList", (rooms) => {
  updateRoomsTable(roomsTable, rooms);
});

socket.on("room/state", (gameRoom) => {
  clearChildren(playerTable);

  const th1 = document.createElement("th");
  const th2 = document.createElement("th");

  appendHeader(playerTable, th1, th2, "Player", "Socket Id");

  gameRoom.players.forEach((player, i) => {
    const td1 = document.createElement("td");
    const td2 = document.createElement("td");
    appendRow(playerTable, td1, td2, i + 1, player);
  });

  currentRoom = gameRoom.name;
  roomName.innerText = gameRoom.name;
  errorInfo.innerText = "";
});

socket.on("room/leave", (rooms) => {
  currentRoom = "";
  leaveRoom();
  updateRoomsTable(roomsTable, rooms);
});

socket.on("room/gameStarted", ({ gameState }) => {
  if (gameState.started) {
    errorInfo.innerText = "";
    startBtn.disabled = true;
    gameTimer.setTime(gameState.timer);
    gameTimer.start();
  }
});

socket.on("room/resume", ({ gameState }) => {
  show(stopBtn);
  hide(resumeBtn);
  gameTimer.setTime(gameState.timer);
  gameTimer.start();
});

socket.on("room/stop", () => {
  hide(stopBtn);
  show(resumeBtn);
  gameTimer.pause();
});

socket.on("room/gameReset", ({ gameState }) => {
  startBtn.disabled = false;
  gameTimer.setTime(gameState.timer);
  gameTimer.pause();
});

socket.on("room/error", (errorMsg) => {
  errorInfo.innerText = errorMsg;
});

socket.on("user/enteredRoom", ({ id }) => {
  playerId.innerText = `Your ID: ${id}`;
  joinRoom();
});

socket.on("room/role", ({ role, location }) => {
  if (role.toLowerCase() == "spy") {
    info.innerText = `your role is Spy`;
  } else {
    info.innerText = `your role is ${role} the location is ${location}`;
  }
});
