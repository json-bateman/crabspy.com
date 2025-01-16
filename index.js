// Set the theme
const storedTheme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "luxury" : "bumblebee";
document.documentElement.setAttribute('data-theme', storedTheme);

// Server
//const socket = io('wss://crabspy.com');
// Testing
const socket = io('ws://localhost:55577');


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
const errorState = document.getElementById('error-state');
const gameState = document.getElementById('game-state');
const changeName = document.getElementById('change-name');
const nameInput = document.getElementById('name-input');

let gameStates = {};

// Log connection status
socket.on('connect', () => {
  console.log('Connected to WebSocket server');
  document.getElementById('info').innerText = 'Connected to server';
});

// Current room name
let currentRoom = '';

// Function to log messages
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

  if (gameStates[roomName]?.gameState.started) {
    errorState.innerText = "Cannot join room, game is in progress"
    return
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

socket.on("room/state", (state) => {
  console.log(state)
  state.players.forEach((i, player) => {
    const tr = document.createElement("tr");
    const td1 = document.createElement("td");
    const td2 = document.createElement("td");
    td1.textContent = i + 1;
    td2.textContent = player;

    tr.appendChild(td2)
    tr.appendChild(td1)
    playerTable.appendChild(tr);
  });
  roomName.innerText = state.roomName
  errorState.innerText = ""
});

socket.on("room/gameStarted", (state) => {
  console.log(state)
  if (state.gameState.started) {
    errorState.innerText = ""
    startBtn.disabled = true;
    join.disabled = true;
  }
});

socket.on("room/error", (state) => {
  console.log(state)
  errorState.innerText = state
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

socket.on("room/timer", ({ remaining }) => {
  timer.innerText = remaining
})

socket.on("room/gameOver", (message) => {
  console.log(message)
  gameState.innerText = message
})
