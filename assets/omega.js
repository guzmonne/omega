(function(){
  function Omega() {}

  Omega.prototype.send = function send(message) {
    console.info(JSON.stringify({
      type: "message",
      message: `${message}`
    }))
  }

  Omega.prototype.start = function start() {
    console.info(JSON.stringify({
      type: "command",
      action: "start",
      message: "Start recording",
    }))
  }

  Omega.prototype.stop = function stop() {
    console.info(JSON.stringify({
      type: "command",
      action: "stop",
      message: "Stop recording",
    }))
  }

  Omega.prototype.done = function done() {
    console.info(JSON.stringify({
      type: "command",
      action: "done",
      message: "Done",
    }))
  }

  Omega.prototype.screenshot = function screenshot() {
    console.info(JSON.stringify({
      type: "command",
      action: "screenshot",
      message: "Take screenshot",
    }))
  }

  // Ready function
  Omega.prototype.ready = function ready(fn) {
    if (document.readyState != 'loading'){
      fn();
    } else {
      document.addEventListener('DOMContentLoaded', fn);
    }
  }
  // Add a new Omega object to the global scope
  window.Omega = new Omega()
})(window)