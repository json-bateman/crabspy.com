function secondsToTimeString(seconds) {
  const minutes = Math.floor((seconds % 3600) / 60);
  const remainingSeconds = Math.floor(seconds % 60);

  const formattedMinutes = minutes.toString()
  const formattedSeconds = remainingSeconds.toString().padStart(2, '0');

  return `${formattedMinutes}:${formattedSeconds}`;
}

function clearChildren(parentEl) {
  while (parentEl.firstChild) {
    parentEl.removeChild(parentEl.firstChild)
  }
}

function appendRow(tParent, el1, el2, c1, c2) {
  const tr = document.createElement("tr");
  el1.textContent = c1;
  el2.textContent = c2;
  tr.appendChild(el1)
  tr.appendChild(el2)
  tParent.appendChild(tr)
}

// Server
const socket = io('wss://crabspy.com');
// Testing
//const socket = io('ws://localhost:55577');


// Grab all the elements, jank style
const join = document.getElementById("join-room")
const roomId = document.getElementById("room-id")
const startBtn = document.getElementById("start-btn")
const resetBtn = document.getElementById("reset-btn")
const player = document.getElementById('player');
const info = document.getElementById('info');
const timer = document.getElementById('timer');
const playerTable = document.getElementById('player-table');
const playersList = document.getElementById('players-list');
const roomName = document.getElementById('room-name');
const errorInfo = document.getElementById('error-info');
const gameInfo = document.getElementById('game-info');
const changeName = document.getElementById('change-name');
const nameInput = document.getElementById('name-input');

let gameStates = {};

// Log connection status
socket.on('connect', () => {
  console.log('Connected to WebSocket server');
  document.getElementById('info').innerText = 'Connected to server';
});

let currentRoom = '';

function infoMessage(message) {
  info.innerText = message;
}

// ~~~~~~~~~~~~~~~~~~~
// All event listeners
// ~~~~~~~~~~~~~~~~~~~
join.addEventListener('click', () => {
  const roomName = roomId.value.trim();
  if (!roomName) {
    infoMessage('Please enter a room name.');
    return;
  }

  currentRoom = roomName;
  socket.emit('room/join', currentRoom);
});

startBtn.addEventListener('click', () => {
  socket.emit('room/start', currentRoom);
});

resetBtn.addEventListener('click', () => {
  socket.emit('room/reset', currentRoom);
});

// ~~~~~~~~~~~~~~~~~~~
// All socket events
// ~~~~~~~~~~~~~~~~~~~
socket.on("allGameStates", (states) => {
  gameStates = states
})

socket.on("room/state", (gameRoom) => {
  console.log("state")
  console.log(gameRoom)

  clearChildren(playerTable);

  const th1 = document.createElement("th");
  const th2 = document.createElement("th");
  appendRow(playerTable, th1, th2, "Player", "Socket Id")

  gameRoom.players.forEach((player, i) => {
    const td1 = document.createElement("td");
    const td2 = document.createElement("td");
    appendRow(playerTable, td1, td2, i + 1, player)
  });

  roomName.innerText = gameRoom.name
  errorInfo.innerText = ""
});

socket.on("room/gameStarted", ({ gameState }) => {
  if (gameState.started) {
    errorInfo.innerText = ""
    startBtn.disabled = true;
    join.disabled = true;
  }
});

socket.on("room/gameReset", ({ gameState }) => {
  timer.innerText = secondsToTimeString(gameState.timer.remaining);
  startBtn.disabled = false;
  join.disabled = false;
})

socket.on("room/error", (errorMsg) => {
  errorInfo.innerText = errorMsg
});

socket.on("room/userInfo", ({ id }) => {
  info.innerText = `Your ID: ${id}`;
});

socket.on("room/role", ({ role, location }) => {
  if (role.toLowerCase() == 'spy') {
    info.innerText = `your role is Spy`
  } else {
    info.innerText = `your role is ${role} the location is ${location}`
  }
})

socket.on("room/timer", ({ gameState }) => {
  timer.innerText = secondsToTimeString(gameState.timer.remaining);
})
