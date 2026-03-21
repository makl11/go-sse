const log = document.getElementById("log")

const eventSource = new EventSource("/sse");
eventSource.onerror = (err) => console.error("ERROR", err);
eventSource.onopen = (event) => console.log("OPEN", event);
eventSource.onmessage = (event) => {
  console.log("MSG", event);
  const entry = createLogEntry(event)
    log.append(entry)
};

/** @param {MessageEvent} event */
function createLogEntry(event) {
    const li = document.createElement("li");
    const data = event.data.startsWith("data: ") ? event.data.slice(6) : event.data;
    li.textContent = data;
    return li;
}
