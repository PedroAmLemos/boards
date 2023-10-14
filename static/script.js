
const canvas = document.getElementById('drawingCanvas');
const ctx = canvas.getContext('2d');
let isDrawing = false;
let lines = [];
let selectedLineIndex = -1;
let selectedPoint = null;
const loc = window.location;

const wsURL = (loc.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + loc.host + '/ws';
const ws = new WebSocket(wsURL);

ws.onmessage = (event) => {
  const line = JSON.parse(event.data);
  lines.push(line);
  redraw();
};

canvas.addEventListener('mousedown', (e) => {
  const rect = canvas.getBoundingClientRect();
  const x = e.clientX - rect.left;
  const y = e.clientY - rect.top;

  lines.some((line, index) => {
    if (isNear({ x, y }, { x: line.x1, y: line.y1 }) || isNear({ x, y }, { x: line.x2, y: line.y2 })) {
      selectedLineIndex = index;
      selectedPoint = isNear({ x, y }, { x: line.x1, y: line.y1 }) ? 'start' : 'end';
      return true;
    }
    return false;
  });

  if (selectedLineIndex === -1) {
    lines.push({ x1: x, y1: y, x2: x, y2: y });
    isDrawing = true;
  }
});

canvas.addEventListener('mousemove', (e) => {
  const rect = canvas.getBoundingClientRect();
  const x = e.clientX - rect.left;
  const y = e.clientY - rect.top;

  if (isDrawing) {
    const currentLine = lines[lines.length - 1];
    currentLine.x2 = x;
    currentLine.y2 = y;
  } else if (selectedLineIndex !== -1) {
    const selectedLine = lines[selectedLineIndex];
    if (selectedPoint === 'start') {
      selectedLine.x1 = x;
      selectedLine.y1 = y;
    } else {
      selectedLine.x2 = x;
      selectedLine.y2 = y;
    }
  } else {
    return;
  }

  redraw();
});

canvas.addEventListener('mouseup', () => {
  if (isDrawing) {
    ws.send(JSON.stringify({
      action: 'create',
      line: lines[lines.length - 1],
    }))
  } else if (selectedLineIndex !== -1) {
    ws.send(JSON.stringify({
      action: 'update',
      index: selectedLineIndex,
      line: lines[selectedLineIndex],
    }))
  }
  isDrawing = false;
  selectedLineIndex = -1;
  selectedPoint = null;
  console.log(lines.length)
});

function redraw() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  lines.forEach(line => {
    ctx.beginPath();
    ctx.moveTo(line.x1, line.y1);
    ctx.lineTo(line.x2, line.y2);
    ctx.stroke();

    ctx.beginPath();
    ctx.arc(line.x1, line.y1, 5, 0, Math.PI * 2);
    ctx.arc(line.x2, line.y2, 5, 0, Math.PI * 2);
    ctx.fillStyle = 'blue';
    ctx.fill();
  });
}

function isNear(point1, point2) {
  const distance = Math.hypot(point1.x - point2.x, point1.y - point2.y);
  return distance < 10;
}
