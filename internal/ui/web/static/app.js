import * as THREE from "three";

const canvas = document.getElementById("scene");
const statsGrid = document.getElementById("stats-grid");
const overlay = document.getElementById("overlay");
const overlayTitle = document.getElementById("overlay-title");
const overlayCopy = document.getElementById("overlay-copy");
const presetSelect = document.getElementById("preset-select");
const startBtn = document.getElementById("start-btn");
const pauseBtn = document.getElementById("pause-btn");
const restartBtn = document.getElementById("restart-btn");
const statusText = document.getElementById("status-text");
const statusBar = document.querySelector(".status-bar");

const renderer = new THREE.WebGLRenderer({ canvas, antialias: true, alpha: true });
renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
renderer.setSize(window.innerWidth, window.innerHeight);
renderer.setScissorTest(true);

const scene = new THREE.Scene();
scene.fog = new THREE.Fog(0x08131e, 18, 40);

const camera = new THREE.OrthographicCamera(
  (-18 * window.innerWidth) / window.innerHeight,
  (18 * window.innerWidth) / window.innerHeight,
  18,
  -18,
  0.1,
  100,
);
camera.position.set(0, 24, 0);
camera.up.set(0, 0, -1);
camera.lookAt(0, 0, 0);

const ambient = new THREE.AmbientLight(0xa7d8ff, 1.7);
scene.add(ambient);

const keyLight = new THREE.DirectionalLight(0xffffff, 1.6);
keyLight.position.set(8, 16, 10);
scene.add(keyLight);

const rimLight = new THREE.DirectionalLight(0x7dffb3, 0.9);
rimLight.position.set(-10, 12, -8);
scene.add(rimLight);

const boardGroup = new THREE.Group();
const snakeGroup = new THREE.Group();
const obstacleGroup = new THREE.Group();
const foodGroup = new THREE.Group();
scene.add(boardGroup, snakeGroup, obstacleGroup, foodGroup);

const boardMaterial = new THREE.MeshStandardMaterial({ color: 0x122537, metalness: 0.08, roughness: 0.78 });
const gridMaterial = new THREE.LineBasicMaterial({ color: 0x31556d, transparent: true, opacity: 0.6 });
const snakeHeadMaterial = new THREE.MeshStandardMaterial({ color: 0x89ff9d, emissive: 0x0a4014, roughness: 0.3 });
const snakeBodyMaterial = new THREE.MeshStandardMaterial({ color: 0x33d37b, emissive: 0x082313, roughness: 0.4 });
const obstacleMaterial = new THREE.MeshStandardMaterial({ color: 0x95a6b5, roughness: 0.75 });
const foodMaterial = new THREE.MeshStandardMaterial({ color: 0xff7c68, emissive: 0x64190b, roughness: 0.2 });

let state = null;
let foodMesh = null;
let lastBoardSize = { width: 40, height: 40 };
let sceneViewport = { left: 0, top: 0, size: 0 };

function buildBoard(width, height) {
  boardGroup.clear();

  const plane = new THREE.Mesh(
    new THREE.BoxGeometry(width + 1, 0.5, height + 1),
    boardMaterial,
  );
  plane.position.set(0, -0.4, 0);
  boardGroup.add(plane);

  for (let x = -width / 2; x <= width / 2; x += 1) {
    const points = [new THREE.Vector3(x, -0.1, -height / 2), new THREE.Vector3(x, -0.1, height / 2)];
    boardGroup.add(new THREE.Line(new THREE.BufferGeometry().setFromPoints(points), gridMaterial));
  }

  for (let z = -height / 2; z <= height / 2; z += 1) {
    const points = [new THREE.Vector3(-width / 2, -0.1, z), new THREE.Vector3(width / 2, -0.1, z)];
    boardGroup.add(new THREE.Line(new THREE.BufferGeometry().setFromPoints(points), gridMaterial));
  }
}

function cellPosition(x, y, width, height, lift = 0.5) {
  return new THREE.Vector3(
    x - width / 2 + 0.5,
    lift,
    y - height / 2 + 0.5,
  );
}

function fillGroup(group, items, width, height, geometryFactory, materialFactory) {
  group.clear();
  items.forEach((item, index) => {
    const mesh = new THREE.Mesh(geometryFactory(index), materialFactory(index));
    mesh.position.copy(cellPosition(item.x, item.y, width, height));
    group.add(mesh);
  });
}

function renderSnapshot(snapshot) {
  if (!snapshot.width || !snapshot.height) {
    return;
  }

  lastBoardSize = { width: snapshot.width, height: snapshot.height };
  buildBoard(snapshot.width, snapshot.height);
  updateCamera(snapshot.width, snapshot.height);

  fillGroup(
    snakeGroup,
    snapshot.snake,
    snapshot.width,
    snapshot.height,
    (index) => new THREE.BoxGeometry(index === 0 ? 0.92 : 0.82, index === 0 ? 0.92 : 0.7, index === 0 ? 0.92 : 0.82),
    (index) => (index === 0 ? snakeHeadMaterial : snakeBodyMaterial),
  );

  fillGroup(
    obstacleGroup,
    snapshot.obstacles,
    snapshot.width,
    snapshot.height,
    () => new THREE.BoxGeometry(0.84, 0.84, 0.84),
    () => obstacleMaterial,
  );

  foodGroup.clear();
  foodMesh = new THREE.Mesh(new THREE.SphereGeometry(0.34, 24, 24), foodMaterial);
  foodMesh.position.copy(cellPosition(snapshot.food.x, snapshot.food.y, snapshot.width, snapshot.height, 0.6));
  foodGroup.add(foodMesh);
}

function stat(label, value) {
  return `<div class="stat"><strong>${label}</strong><span>${value}</span></div>`;
}

function formatDuration(ms) {
  const total = Math.max(0, Math.floor(ms / 1000));
  const minutes = String(Math.floor(total / 60)).padStart(2, "0");
  const seconds = String(total % 60).padStart(2, "0");
  return `${minutes}:${seconds}`;
}

function updateHUD(payload) {
  const { snapshot } = payload;

  statsGrid.innerHTML = [
    stat("Score", snapshot.score),
    stat("Length", snapshot.snake.length),
    stat("Level", snapshot.level || 1),
    stat("Food", snapshot.food_eaten),
    stat("Next Level", snapshot.foods_to_next_level),
    stat("Obstacles", snapshot.obstacles.length),
    stat("Elapsed", formatDuration(snapshot.elapsed_millis)),
    stat("Speed", `${snapshot.tick_interval_millis ? (1000 / snapshot.tick_interval_millis).toFixed(1) : "0.0"}/s`),
  ].join("");

  if (!snapshot.width) {
    overlay.classList.remove("hidden");
    overlayTitle.textContent = "Start a Run";
    overlayCopy.textContent = "Pick a preset and press Start. The first movement key begins the run.";
  } else if (!snapshot.started && !snapshot.is_over) {
    overlay.classList.remove("hidden");
    overlayTitle.textContent = "Ready";
    overlayCopy.textContent = "Use WASD or the arrow keys to launch the snake.";
  } else if (snapshot.is_over) {
    overlay.classList.remove("hidden");
    overlayTitle.textContent = snapshot.is_won ? "You Win" : "Game Over";
    overlayCopy.textContent = snapshot.has_last_run
      ? `Score ${snapshot.last_run.score}, length ${snapshot.last_run.length}, time ${formatDuration(snapshot.last_run.duration_millis)}. Press Restart to run again.`
      : "Press Restart to run again.";
  } else if (snapshot.paused) {
    overlay.classList.remove("hidden");
    overlayTitle.textContent = "Paused";
    overlayCopy.textContent = "Resume with P or continue from the panel controls.";
  } else {
    overlay.classList.add("hidden");
  }

  pauseBtn.textContent = snapshot.paused ? "Resume" : "Pause";
}

async function callJSON(url, body) {
  const response = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body ?? {}),
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(payload.error || "Request failed");
  }
  return payload;
}

async function fetchState() {
  const response = await fetch("/api/state", { cache: "no-store" });
  if (!response.ok) {
    throw new Error("Could not load state");
  }
  return response.json();
}

function setStatus(message, isError = false) {
  statusText.textContent = message;
  statusText.classList.toggle("error", isError);
}

async function refresh() {
  try {
    state = await fetchState();
    if (!presetSelect.options.length) {
      state.presets.forEach((preset, index) => {
        const option = document.createElement("option");
        option.value = String(index);
        option.textContent = `${index + 1}. ${preset.name}`;
        presetSelect.appendChild(option);
      });
    }
    presetSelect.value = String(state.current_preset);
    renderSnapshot(state.snapshot);
    updateHUD(state);
    setStatus("Connected");
  } catch (error) {
    setStatus(error.message, true);
  }
}

function animate() {
  requestAnimationFrame(animate);
  const time = performance.now() * 0.001;
  if (foodMesh) {
    foodMesh.rotation.y = time * 1.6;
    foodMesh.rotation.x = 0.35 + Math.sin(time * 1.4) * 0.08;
  }
  renderer.setViewport(
    sceneViewport.left,
    window.innerHeight - sceneViewport.top - sceneViewport.size,
    sceneViewport.size,
    sceneViewport.size,
  );
  renderer.setScissor(
    sceneViewport.left,
    window.innerHeight - sceneViewport.top - sceneViewport.size,
    sceneViewport.size,
    sceneViewport.size,
  );
  renderer.render(scene, camera);
}

function updateCamera(boardWidth, boardHeight) {
  const halfBoard = Math.max(boardWidth, boardHeight) / 2 + 1.2;

  camera.left = -halfBoard;
  camera.right = halfBoard;
  camera.top = halfBoard;
  camera.bottom = -halfBoard;

  camera.position.set(0, 24, 0);
  camera.lookAt(0, 0, 0);
  camera.updateProjectionMatrix();
}

function updateViewport() {
  const barRect = statusBar?.getBoundingClientRect();
  const topOffset = (barRect?.bottom ?? 0) + 18;
  const availableWidth = Math.max(200, window.innerWidth - 32);
  const availableHeight = Math.max(200, window.innerHeight - topOffset - 18);
  const size = Math.max(200, Math.min(availableWidth, availableHeight));

  sceneViewport = {
    left: Math.round((window.innerWidth - size) / 2),
    top: Math.round(topOffset + (availableHeight - size) / 2),
    size,
  };
}

window.addEventListener("resize", () => {
  renderer.setSize(window.innerWidth, window.innerHeight);
  updateViewport();
  updateCamera(lastBoardSize.width, lastBoardSize.height);
});

window.addEventListener("keydown", async (event) => {
  const key = event.key.toLowerCase();
  const directions = {
    arrowup: "up",
    w: "up",
    arrowdown: "down",
    s: "down",
    arrowleft: "left",
    a: "left",
    arrowright: "right",
    d: "right",
  };

  try {
    if (directions[key]) {
      await callJSON("/api/input", { direction: directions[key] });
      await refresh();
      return;
    }
    if (key === "p") {
      await callJSON("/api/pause");
      await refresh();
      return;
    }
    if (key === "r") {
      await callJSON("/api/restart");
      await refresh();
      return;
    }
    if (key === "enter" && !state?.snapshot?.width) {
      await callJSON("/api/start", { preset: Number(presetSelect.value) });
      await refresh();
    }
  } catch (error) {
    setStatus(error.message, true);
  }
});

startBtn.addEventListener("click", async () => {
  try {
    await callJSON("/api/start", { preset: Number(presetSelect.value) });
    await refresh();
  } catch (error) {
    setStatus(error.message, true);
  }
});

pauseBtn.addEventListener("click", async () => {
  try {
    await callJSON("/api/pause");
    await refresh();
  } catch (error) {
    setStatus(error.message, true);
  }
});

restartBtn.addEventListener("click", async () => {
  try {
    await callJSON("/api/restart");
    await refresh();
  } catch (error) {
    setStatus(error.message, true);
  }
});

refresh();
updateViewport();
setInterval(refresh, 120);
animate();
