const http = require("http").createServer();
const fs = require("fs")
const locationData = require('../locations.json')
const locations = locationData.locations

const io = require("socket.io")(http, {
  cors: { origin: "*" }
});

// APP STATE
let gameRooms = {
  roomName: {
    players: [],
    names: {},
    gameState: {
      location: null,
      roles: {},
      spy: null,
      timer: { startTime: null, duration: 0 },
    },
  },
};

io.on("connection", (socket) => {
  console.log('A user connected', socket.id);

  socket.on("room/join", (roomName) => {
    console.log(`User ${socket.id} joined room: ${roomName}`);

    if (!gameRooms[roomName]) {
      gameRooms[roomName] = { players: [], names: {} };
    }

    // Add the player and set default name as their socket ID
    if (!gameRooms[roomName].players.includes(socket.id)) {
      gameRooms[roomName].players.push(socket.id);
      gameRooms[roomName].names[socket.id] = socket.id;
    } else {
      return
    }

    socket.join(roomName);

    io.to(roomName).emit("room/state", {
      players: gameRooms[roomName].players,
      names: gameRooms[roomName].names,
      room: roomName
    });

    socket.emit("room/userInfo", {
      id: socket.id,
    });
  });

  socket.on("room/start", (roomName) => {
    if (!gameRooms[roomName] || gameRooms[roomName].players.length < 2) {
      console.log("Not enough players to start the game");
      io.to(roomName).emit("room/error", "Not enough players to start the game.");
      return;
    }

    const players = [...gameRooms[roomName].players];
    const randomIndex = Math.floor(Math.random() * locations.length);
    const selectedLocation = locations[randomIndex];
    const locationRoles = [...selectedLocation.roles];

    // Shuffle players
    players.sort(() => Math.random() - 0.5);

    // Select a spy
    const spyIndex = Math.floor(Math.random() * players.length);
    const spyId = players[spyIndex];

    // Remove spy from role assignment pool
    const nonSpyPlayers = players.filter(player => player !== spyId);

    // Shuffle roles
    locationRoles.sort(() => Math.random() - 0.5);

    // Assign roles to non-spy players
    const roleAssignments = {};
    nonSpyPlayers.forEach((playerId, index) => {
      roleAssignments[playerId] = locationRoles[index % locationRoles.length];
    });

    // Store the game state
    const duration = 300;
    const startTime = Date.now();
    gameRooms[roomName].gameState = {
      location: selectedLocation.title,
      roles: roleAssignments,
      spy: spyId,
      timer: { startTime, duration },
    };

    console.log(`Game started in room: ${roomName}`);
    console.log(`Location: ${selectedLocation.title}`);
    console.log(`Roles:`, roleAssignments);
    console.log(`Spy: ${spyId}`);

    // Notify players of their roles
    players.forEach(playerId => {
      const role = playerId === spyId ? "Spy" : roleAssignments[playerId];
      io.to(playerId).emit("room/role", { role, location: playerId === spyId ? null : selectedLocation.title });
    });

    io.to(roomName).emit("room/gameStarted", {
      message: `Game has started in: ${selectedLocation.title}`,
      location: selectedLocation.title,
      duration,
    });
  });

  socket.on("room/reset", (roomName) => {
    console.log(`Game reset in room: ${roomName}`);
    gameRooms[roomName] = { players: [] };
    io.to(roomName).emit("room/state", gameRooms[roomName]);
  });



  // Handle disconnection
  socket.on("disconnect", () => {
    console.log(`User ${socket.id} disconnected`);
    // Remove the user from all rooms
    for (const roomName in gameRooms) {
      gameRooms[roomName].players = gameRooms[roomName].players.filter(id => id !== socket.id);
      io.to(roomName).emit("room/state", gameRooms[roomName]);
    }
  });
});

setInterval(() => {
  for (const roomName in gameRooms) {
    const gameState = gameRooms[roomName]?.gameState;
    if (gameState?.timer?.startTime) {
      const elapsed = Math.floor((Date.now() - gameState.timer.startTime) / 1000);
      const remaining = Math.max(0, gameState.timer.duration - elapsed);

      // Broadcast timer to the room
      io.to(roomName).emit("room/timer", { remaining });

      // End the game if time is up
      if (remaining <= 0) {
        io.to(roomName).emit("room/gameOver", "Time's up! The game is over.");
        delete gameRooms[roomName].gameState.timer; // Reset the timer
      }
    }
  }
}, 1000);

http.listen(55577, () => console.log('Listening on port 55577'));
