const http = require("http").createServer();
const locationData = require('../locations.json')
const locations = locationData.locations

const io = require("socket.io")(http, {
  cors: { origin: "*" }
});

// APP STATE
const gameRooms = {
  _defaultRoom: {
    roomName: "",
    players: [],
    gameState: {
      started: false,
      location: null,
      roles: {},
      spy: null,
      timer: { startTime: null, duration: 0 },
    },
  },
};

io.on("connection", (socket) => {
  console.log('A user connected', socket.id);

  // When user hits website, give them the state of all games
  socket.emit('allGameStates', gameRooms)

  socket.on("room/join", (room) => {
    console.log(`User ${socket.id} joined room: ${room}`);

    if (!gameRooms[room]) {
      gameRooms[room] = gameRooms._defaultRoom;
      gameRooms[room].roomName = room;
    }

    // Add the player and set default name as their socket ID
    if (!gameRooms[room].players.includes(socket.id)) {
      gameRooms[room].players.push(socket.id);
    } else {
      return
    }

    socket.join(room);

    io.to(room).emit("room/state", gameRooms[room]);

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
    const duration = 5;
    const startTime = Date.now();
    gameRooms[roomName].gameState = {
      location: selectedLocation.title,
      roles: roleAssignments,
      spy: spyId,
      timer: { startTime, duration },
      started: true,
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

    io.to(roomName).emit("room/gameStarted", gameRooms[roomName]);
  });

  socket.on("room/reset", (roomName) => {
    gameRooms[roomName] = gameRooms._defaultRoom
    gameRooms[roomName].roomName = roomName
  })

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
    console.log(gameRooms)
    const gameState = gameRooms[roomName].gameState;
    if (gameState?.timer?.startTime) {
      const elapsed = Math.floor((Date.now() - gameState.timer.startTime) / 1000);
      const remaining = Math.max(0, gameState.timer.duration - elapsed);

      io.to(roomName).emit("room/timer", { remaining });

      if (remaining <= 0) {
        console.log("here")
        io.to(roomName).emit("room/gameOver", "Time's up! The game is over.");
        delete gameRooms[roomName].gameState.timer;
        gameRooms[roomName].gameState.started = false
        io.to(roomName).emit("room/state", gameRooms[roomName]);
      }
    }
  }
}, 1000);

http.listen(55577, () => console.log('Listening on port 55577'));
