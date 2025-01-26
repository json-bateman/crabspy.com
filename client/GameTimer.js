class GameTimer {
  constructor(displayElement) {
    this.displayElement = displayElement;
    this.timeRemaining = 0;
    this.intervalId = null;
    this.isPaused = true;
  }

  setTime(seconds) {
    this.timeRemaining = seconds;
    this.updateDisplay();
  }

  start() {
    if (this.isPaused) {
      this.isPaused = false;
      this.tick();
      this.intervalId = setInterval(() => this.tick(), 1000);
    }
  }

  pause() {
    this.isPaused = true;
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
  }

  tick() {
    if (!this.isPaused && this.timeRemaining > 0) {
      this.timeRemaining--;
      this.updateDisplay();
    }
  }

  updateDisplay() {
    const minutes = Math.floor(this.timeRemaining / 60);
    const seconds = this.timeRemaining % 60;
    const formattedSeconds = seconds.toString().padStart(2, "0");
    this.displayElement.textContent = `${minutes}:${formattedSeconds}`;
  }
}
