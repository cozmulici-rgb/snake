import * as THREE from "three";

const canvas = document.getElementById("scene");
const statsGrid = document.getElementById("stats-grid");
const scoreCard = document.getElementById("score-card");
const levelValue = document.getElementById("level-value");
const progressLabel = document.getElementById("progress-label");
const progressMeta = document.getElementById("progress-meta");
const progressFill = document.getElementById("progress-fill");
const progressCaption = document.getElementById("progress-caption");
const overlay = document.getElementById("overlay");
const overlayTitle = document.getElementById("overlay-title");
const overlayCopy = document.getElementById("overlay-copy");
const presetSelect = document.getElementById("preset-select");
const startBtn = document.getElementById("start-btn");
const pauseBtn = document.getElementById("pause-btn");
const restartBtn = document.getElementById("restart-btn");
const developerModeToggle = document.getElementById("developer-mode-toggle");
const developerLevelInput = document.getElementById("developer-level-input");
const developerLevelBtn = document.getElementById("developer-level-btn");
const developerHint = document.getElementById("developer-hint");
const statusText = document.getElementById("status-text");
const statusBar = document.querySelector(".status-bar");
const layoutGap = 18;
const narrowBreakpoint = 1100;

const renderer = new THREE.WebGLRenderer({ canvas, antialias: true, alpha: true });
renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
renderer.setSize(window.innerWidth, window.innerHeight);
renderer.setScissorTest(true);
renderer.outputColorSpace = THREE.SRGBColorSpace;
renderer.toneMapping = THREE.ACESFilmicToneMapping;
renderer.toneMappingExposure = 1.08;

const scene = new THREE.Scene();
scene.fog = new THREE.Fog(0x07111b, 28, 58);

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

const ambient = new THREE.AmbientLight(0xa7d8ff, 1.2);
scene.add(ambient);

const hemisphere = new THREE.HemisphereLight(0xa8deff, 0x03070d, 1.05);
scene.add(hemisphere);

const keyLight = new THREE.DirectionalLight(0xf8fbff, 1.35);
keyLight.position.set(9, 18, 12);
scene.add(keyLight);

const rimLight = new THREE.DirectionalLight(0x7dffb3, 0.7);
rimLight.position.set(-12, 10, -10);
scene.add(rimLight);

const warmFillLight = new THREE.DirectionalLight(0xff9d72, 0.45);
warmFillLight.position.set(7, 10, -6);
scene.add(warmFillLight);

const boardGroup = new THREE.Group();
const snakeShadowGroup = new THREE.Group();
const obstacleShadowGroup = new THREE.Group();
const foodShadowGroup = new THREE.Group();
const snakeGlowGroup = new THREE.Group();
const snakeGroup = new THREE.Group();
const headMarkerGroup = new THREE.Group();
const obstacleGlowGroup = new THREE.Group();
const obstacleGroup = new THREE.Group();
const foodGlowGroup = new THREE.Group();
const foodGroup = new THREE.Group();
scene.add(
  boardGroup,
  snakeShadowGroup,
  obstacleShadowGroup,
  foodShadowGroup,
  snakeGlowGroup,
  snakeGroup,
  headMarkerGroup,
  obstacleGlowGroup,
  obstacleGroup,
  foodGlowGroup,
  foodGroup,
);

const boardMaterial = new THREE.MeshStandardMaterial({
  color: 0x0c1724,
  emissive: 0x07111a,
  emissiveIntensity: 0.75,
  metalness: 0.04,
  roughness: 0.92,
});
const boardAccentMaterial = new THREE.MeshBasicMaterial({
  color: 0x194264,
  transparent: true,
  opacity: 0.09,
  depthWrite: false,
  blending: THREE.AdditiveBlending,
});
const boardAccentOuterMaterial = new THREE.MeshBasicMaterial({
  color: 0x10263c,
  transparent: true,
  opacity: 0.12,
  depthWrite: false,
});
const gridMaterial = new THREE.LineBasicMaterial({ color: 0x84a8c2, transparent: true, opacity: 0.11 });
const majorGridMaterial = new THREE.LineBasicMaterial({ color: 0x8fb3d0, transparent: true, opacity: 0.17 });
const boardFrameMaterial = new THREE.LineBasicMaterial({ color: 0xa9d5ff, transparent: true, opacity: 0.24 });
const snakeHeadMaterial = new THREE.MeshStandardMaterial({
  color: 0xcaffbc,
  emissive: 0x2dff74,
  emissiveIntensity: 1.1,
  roughness: 0.2,
  metalness: 0.06,
});
const snakeBodyMaterial = new THREE.MeshStandardMaterial({
  color: 0x36dc80,
  emissive: 0x0d542f,
  emissiveIntensity: 0.8,
  roughness: 0.34,
  metalness: 0.04,
});
const obstacleMaterial = new THREE.MeshStandardMaterial({
  color: 0x6d83ff,
  emissive: 0x182d82,
  emissiveIntensity: 0.8,
  roughness: 0.44,
  metalness: 0.12,
});
const foodMaterial = new THREE.MeshStandardMaterial({
  color: 0xffaf64,
  emissive: 0xff5d22,
  emissiveIntensity: 1.35,
  roughness: 0.14,
  metalness: 0.05,
});
const snakeHeadGlowMaterial = new THREE.MeshBasicMaterial({
  color: 0x63ffa2,
  transparent: true,
  opacity: 0.28,
  depthWrite: false,
  blending: THREE.AdditiveBlending,
});
const snakeBodyGlowMaterial = new THREE.MeshBasicMaterial({
  color: 0x36dc80,
  transparent: true,
  opacity: 0.14,
  depthWrite: false,
  blending: THREE.AdditiveBlending,
});
const obstacleGlowMaterial = new THREE.MeshBasicMaterial({
  color: 0x667cff,
  transparent: true,
  opacity: 0.18,
  depthWrite: false,
  blending: THREE.AdditiveBlending,
});
const foodGlowMaterial = new THREE.MeshBasicMaterial({
  color: 0xff7d45,
  transparent: true,
  opacity: 0.3,
  depthWrite: false,
  blending: THREE.AdditiveBlending,
});
const snakeHeadShadowMaterial = new THREE.MeshBasicMaterial({ color: 0x041109, transparent: true, opacity: 0.28, depthWrite: false });
const snakeBodyShadowMaterial = new THREE.MeshBasicMaterial({ color: 0x041109, transparent: true, opacity: 0.16, depthWrite: false });
const obstacleShadowMaterial = new THREE.MeshBasicMaterial({ color: 0x040814, transparent: true, opacity: 0.18, depthWrite: false });
const foodShadowMaterial = new THREE.MeshBasicMaterial({ color: 0x140703, transparent: true, opacity: 0.18, depthWrite: false });
const headMarkerMaterial = new THREE.MeshBasicMaterial({ color: 0x0b2415 });
const snakeHeadGeometry = new THREE.BoxGeometry(1.04, 0.96, 1.04);
const snakeBodyGeometry = new THREE.BoxGeometry(0.88, 0.56, 0.88);
const obstacleGeometry = new THREE.BoxGeometry(0.86, 0.78, 0.86);
const foodGeometry = new THREE.SphereGeometry(0.4, 28, 28);
const snakeGlowGeometry = new THREE.CircleGeometry(0.56, 24);
const obstacleGlowGeometry = new THREE.CircleGeometry(0.56, 24);
const foodGlowGeometry = new THREE.CircleGeometry(0.68, 28);
const boardAccentGeometry = new THREE.CircleGeometry(0.5, 48);
const snakeShadowGeometry = new THREE.CircleGeometry(0.52, 24);
const obstacleShadowGeometry = new THREE.CircleGeometry(0.5, 24);
const foodShadowGeometry = new THREE.CircleGeometry(0.48, 24);
const headMarkerGeometry = new THREE.BoxGeometry(0.2, 0.12, 0.36);

let state = null;
let foodMesh = null;
let headMarkerMesh = null;
let lastBoardSize = { width: 40, height: 40 };
let sceneViewport = { left: 0, top: 0, size: 0 };
let lastBoardKey = "";
let lastHUDKey = "";
let refreshTimer = null;
let refreshInFlight = false;

function buildBoard(width, height) {
  const boardKey = `${width}x${height}`;
  if (lastBoardKey === boardKey) {
    return;
  }

  boardGroup.clear();
  lastBoardKey = boardKey;

  const plane = new THREE.Mesh(
    new THREE.BoxGeometry(width + 1, 0.5, height + 1),
    boardMaterial,
  );
  plane.position.set(0, -0.4, 0);
  boardGroup.add(plane);

  const outerAccent = new THREE.Mesh(boardAccentGeometry, boardAccentOuterMaterial);
  outerAccent.rotation.x = -Math.PI / 2;
  outerAccent.position.set(0, -0.14, 0);
  outerAccent.scale.set(width * 0.92, height * 0.92, 1);
  boardGroup.add(outerAccent);

  const innerAccent = new THREE.Mesh(boardAccentGeometry, boardAccentMaterial);
  innerAccent.rotation.x = -Math.PI / 2;
  innerAccent.position.set(0, -0.13, 0);
  innerAccent.scale.set(width * 0.54, height * 0.54, 1);
  boardGroup.add(innerAccent);

  for (let x = -width / 2; x <= width / 2; x += 1) {
    const points = [new THREE.Vector3(x, -0.1, -height / 2), new THREE.Vector3(x, -0.1, height / 2)];
    const material = x % 5 === 0 ? majorGridMaterial : gridMaterial;
    boardGroup.add(new THREE.Line(new THREE.BufferGeometry().setFromPoints(points), material));
  }

  for (let z = -height / 2; z <= height / 2; z += 1) {
    const points = [new THREE.Vector3(-width / 2, -0.1, z), new THREE.Vector3(width / 2, -0.1, z)];
    const material = z % 5 === 0 ? majorGridMaterial : gridMaterial;
    boardGroup.add(new THREE.Line(new THREE.BufferGeometry().setFromPoints(points), material));
  }

  const framePoints = [
    new THREE.Vector3(-width / 2, -0.08, -height / 2),
    new THREE.Vector3(width / 2, -0.08, -height / 2),
    new THREE.Vector3(width / 2, -0.08, height / 2),
    new THREE.Vector3(-width / 2, -0.08, height / 2),
    new THREE.Vector3(-width / 2, -0.08, -height / 2),
  ];
  boardGroup.add(new THREE.Line(new THREE.BufferGeometry().setFromPoints(framePoints), boardFrameMaterial));
}

function cellPosition(x, y, width, height, lift = 0.5) {
  return new THREE.Vector3(
    x - width / 2 + 0.5,
    lift,
    y - height / 2 + 0.5,
  );
}

function ensureMesh(group, index, geometry, material) {
  while (group.children.length <= index) {
    const mesh = new THREE.Mesh(geometry, material);
    group.add(mesh);
  }
  const mesh = group.children[index];
  mesh.visible = true;
  return mesh;
}

function ensureFlatMesh(group, index, geometry, material) {
  const mesh = ensureMesh(group, index, geometry, material);
  mesh.rotation.x = -Math.PI / 2;
  mesh.renderOrder = 1;
  return mesh;
}

function fillGroup(group, items, width, height, geometry, material, materialSelector = null) {
  items.forEach((item, index) => {
    const mesh = ensureMesh(group, index, geometry, materialSelector ? materialSelector(index) : material);
    if (materialSelector) {
      mesh.material = materialSelector(index);
    }
    mesh.position.copy(cellPosition(item.x, item.y, width, height));
  });

  for (let index = items.length; index < group.children.length; index += 1) {
    group.children[index].visible = false;
  }
}

function fillFlatGroup(group, items, width, height, geometry, material, lift = 0.04, materialSelector = null, scaleSelector = null) {
  items.forEach((item, index) => {
    const mesh = ensureFlatMesh(group, index, geometry, materialSelector ? materialSelector(index) : material);
    if (materialSelector) {
      mesh.material = materialSelector(index);
    }
    mesh.position.copy(cellPosition(item.x, item.y, width, height, lift));
    const scale = scaleSelector ? scaleSelector(index, item) : 1;
    mesh.scale.setScalar(scale);
  });

  for (let index = items.length; index < group.children.length; index += 1) {
    group.children[index].visible = false;
  }
}

function headMarkerPose(direction) {
  switch (direction) {
    case "up":
      return { dx: 0, dz: -0.24, rotationY: 0 };
    case "down":
      return { dx: 0, dz: 0.24, rotationY: 0 };
    case "left":
      return { dx: -0.24, dz: 0, rotationY: Math.PI / 2 };
    case "right":
      return { dx: 0.24, dz: 0, rotationY: Math.PI / 2 };
    default:
      return { dx: 0, dz: -0.24, rotationY: 0 };
  }
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
    snakeBodyGeometry,
    snakeBodyMaterial,
    (index) => (index === 0 ? snakeHeadMaterial : snakeBodyMaterial),
  );
  fillFlatGroup(
    snakeShadowGroup,
    snapshot.snake,
    snapshot.width,
    snapshot.height,
    snakeShadowGeometry,
    snakeBodyShadowMaterial,
    0.018,
    (index) => (index === 0 ? snakeHeadShadowMaterial : snakeBodyShadowMaterial),
    (index) => (index === 0 ? 1.04 : 0.94),
  );
  fillFlatGroup(
    snakeGlowGroup,
    snapshot.snake,
    snapshot.width,
    snapshot.height,
    snakeGlowGeometry,
    snakeBodyGlowMaterial,
    0.03,
    (index) => (index === 0 ? snakeHeadGlowMaterial : snakeBodyGlowMaterial),
    (index) => (index === 0 ? 1.3 : 1.05),
  );

  fillGroup(
    obstacleGroup,
    snapshot.obstacles,
    snapshot.width,
    snapshot.height,
    obstacleGeometry,
    obstacleMaterial,
  );
  fillFlatGroup(
    obstacleShadowGroup,
    snapshot.obstacles,
    snapshot.width,
    snapshot.height,
    obstacleShadowGeometry,
    obstacleShadowMaterial,
    0.018,
    null,
    () => 0.96,
  );
  fillFlatGroup(
    obstacleGlowGroup,
    snapshot.obstacles,
    snapshot.width,
    snapshot.height,
    obstacleGlowGeometry,
    obstacleGlowMaterial,
    0.03,
    null,
    () => 1.08,
  );

  if (!foodMesh) {
    foodMesh = new THREE.Mesh(foodGeometry, foodMaterial);
    foodGroup.add(foodMesh);
  }
  foodMesh.visible = true;
  foodMesh.position.copy(cellPosition(snapshot.food.x, snapshot.food.y, snapshot.width, snapshot.height, 0.6));

  fillFlatGroup(
    foodShadowGroup,
    [snapshot.food],
    snapshot.width,
    snapshot.height,
    foodShadowGeometry,
    foodShadowMaterial,
    0.018,
    null,
    () => 1,
  );
  fillFlatGroup(
    foodGlowGroup,
    [snapshot.food],
    snapshot.width,
    snapshot.height,
    foodGlowGeometry,
    foodGlowMaterial,
    0.03,
    null,
    () => 1.1,
  );

  if (!headMarkerMesh) {
    headMarkerMesh = new THREE.Mesh(headMarkerGeometry, headMarkerMaterial);
    headMarkerGroup.add(headMarkerMesh);
  }

  if (snapshot.snake.length > 0) {
    const head = snapshot.snake[0];
    const pose = headMarkerPose(snapshot.direction);
    const headPosition = cellPosition(head.x, head.y, snapshot.width, snapshot.height, 0.96);
    headMarkerMesh.visible = true;
    headMarkerMesh.position.set(headPosition.x + pose.dx, headPosition.y, headPosition.z + pose.dz);
    headMarkerMesh.rotation.y = pose.rotationY;
  } else {
    headMarkerMesh.visible = false;
  }
}

function stat(label, value) {
  return `<div class="stat"><strong>${label}</strong><span>${value}</span></div>`;
}

function scoreMarkup(score, bestScore) {
  const caption = bestScore > score ? `Best ${bestScore}` : (score > 0 ? "Run active" : "Press Start");
  return `
    <span class="eyebrow">Score</span>
    <span class="score-value">${score}</span>
    <span class="score-caption">${caption}</span>
  `;
}

function formatDuration(ms) {
  const total = Math.max(0, Math.floor(ms / 1000));
  const minutes = String(Math.floor(total / 60)).padStart(2, "0");
  const seconds = String(total % 60).padStart(2, "0");
  return `${minutes}:${seconds}`;
}

function foodsPerLevelForState(payload) {
  const preset = payload.presets?.[payload.current_preset];
  return Math.max(1, preset?.foods_per_level ?? 1);
}

function levelProgress(snapshot, foodsPerLevel) {
  const collectedThisLevel = snapshot.food_eaten > 0
    ? (foodsPerLevel - snapshot.foods_to_next_level) % foodsPerLevel
    : 0;
  const normalizedCollected = Math.max(0, Math.min(foodsPerLevel, collectedThisLevel));
  const remaining = Math.max(0, snapshot.foods_to_next_level);
  const ratio = foodsPerLevel > 0 ? normalizedCollected / foodsPerLevel : 0;

  return {
    collected: normalizedCollected,
    total: foodsPerLevel,
    remaining,
    ratio,
  };
}

function updateHUD(payload) {
  const { snapshot } = payload;
  const foodsPerLevel = foodsPerLevelForState(payload);
  const progress = levelProgress(snapshot, foodsPerLevel);
  const hudKey = [
    snapshot.score,
    snapshot.snake.length,
    snapshot.level,
    snapshot.food_eaten,
    snapshot.foods_to_next_level,
    snapshot.obstacles.length,
    snapshot.elapsed_millis,
    snapshot.tick_interval_millis,
    snapshot.best_score,
    snapshot.started,
    snapshot.is_over,
    snapshot.is_won,
    snapshot.paused,
    snapshot.has_last_run,
    snapshot.last_run.score,
    snapshot.last_run.length,
    snapshot.last_run.duration_millis,
    foodsPerLevel,
  ].join("|");

  if (hudKey !== lastHUDKey) {
    scoreCard.innerHTML = scoreMarkup(snapshot.score, snapshot.best_score);
    levelValue.textContent = String(snapshot.level || 1);
    progressLabel.textContent = "Next Level";
    progressMeta.textContent = `${progress.collected} / ${progress.total} food`;
    progressFill.style.width = `${Math.max(0, Math.min(100, progress.ratio * 100))}%`;
    progressCaption.textContent = progress.remaining > 0
      ? `${progress.remaining} food to level ${Math.max(1, (snapshot.level || 1) + 1)}`
      : `Level ${snapshot.level || 1} is ready to advance`;

    const secondaryStats = [
      stat("Length", snapshot.snake.length),
      snapshot.started || snapshot.food_eaten > 0 || snapshot.is_over
        ? stat("Food Eaten", snapshot.food_eaten)
        : "",
      snapshot.started || snapshot.paused || snapshot.is_over
        ? stat("Time", formatDuration(snapshot.elapsed_millis))
        : "",
      snapshot.started || snapshot.paused || snapshot.is_over
        ? stat("Speed", `${snapshot.tick_interval_millis ? (1000 / snapshot.tick_interval_millis).toFixed(1) : "0.0"}/s`)
        : "",
      snapshot.obstacles.length > 0
        ? stat("Obstacles", snapshot.obstacles.length)
        : "",
    ].filter(Boolean);

    statsGrid.innerHTML = secondaryStats.join("");
    lastHUDKey = hudKey;
  }

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
  pauseBtn.hidden = !snapshot.started || snapshot.is_over;
  restartBtn.hidden = !snapshot.is_over && !snapshot.paused;
  startBtn.hidden = snapshot.is_over || snapshot.paused;
  pauseBtn.disabled = !snapshot.width || snapshot.is_over;
  restartBtn.disabled = !snapshot.width;
  startBtn.disabled = snapshot.started && !snapshot.is_over;
  developerModeToggle.checked = Boolean(payload.developer_mode);
  developerModeToggle.disabled = false;
  developerLevelInput.disabled = !payload.developer_mode || !snapshot.width || snapshot.is_over;
  developerLevelBtn.disabled = developerLevelInput.disabled;
  developerHint.textContent = payload.developer_mode
    ? "Developer mode is active. Level jumps keep progression coherent and do not update saved stats."
    : "Enable developer mode to jump the current run to any level. Developer runs do not update saved stats.";
  if (document.activeElement !== developerLevelInput || !Number(developerLevelInput.value) || Number(developerLevelInput.value) < 1) {
    developerLevelInput.value = String(Math.max(1, snapshot.level || 1));
  }
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

function renderGameToText() {
  const snapshot = state?.snapshot ?? null;
  const foodsPerLevel = state ? foodsPerLevelForState(state) : 1;
  const progress = snapshot ? levelProgress(snapshot, foodsPerLevel) : null;
  const payload = {
    mode: overlay.classList.contains("hidden") ? "live" : (overlayTitle.textContent || "overlay").toLowerCase(),
    developer_mode: Boolean(state?.developer_mode),
    viewport: sceneViewport,
    snapshot: snapshot ? {
      width: snapshot.width,
      height: snapshot.height,
      score: snapshot.score,
      level: snapshot.level,
      started: snapshot.started,
      paused: snapshot.paused,
      is_over: snapshot.is_over,
      foods_per_level: foodsPerLevel,
      progress_to_next_level: progress ? {
        collected: progress.collected,
        total: progress.total,
        remaining: progress.remaining,
      } : null,
      snake: snapshot.snake,
      food: snapshot.food,
      obstacles: snapshot.obstacles,
    } : null,
  };
  return JSON.stringify(payload);
}

window.render_game_to_text = renderGameToText;
window.advanceTime = async (ms = 0) => {
  await new Promise((resolve) => window.setTimeout(resolve, Math.max(0, ms)));
  await refresh();
  return renderGameToText();
};

async function refresh() {
  if (refreshInFlight) {
    return;
  }

  refreshInFlight = true;
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
    updateViewport();
    setStatus("Connected");
  } catch (error) {
    setStatus(error.message, true);
  } finally {
    refreshInFlight = false;
    scheduleRefresh();
  }
}

async function updateDeveloperMode(enabled) {
  try {
    await callJSON("/api/developer-mode", { enabled });
    setStatus(enabled ? "Developer mode enabled" : "Developer mode disabled");
    await refresh();
  } catch (error) {
    developerModeToggle.checked = !enabled;
    setStatus(error.message, true);
  }
}

async function bypassLevel(level) {
  try {
    await callJSON("/api/developer-level", { level });
    developerLevelInput.value = String(level);
    setStatus(`Jumped to level ${level}`);
    await refresh();
  } catch (error) {
    setStatus(error.message, true);
  }
}

function nextRefreshDelay() {
  const tick = state?.snapshot?.tick_interval_millis ?? 140;
  return Math.max(33, Math.min(80, Math.floor(tick / 2)));
}

function scheduleRefresh() {
  if (refreshTimer) {
    clearTimeout(refreshTimer);
  }
  refreshTimer = setTimeout(() => {
    refresh();
  }, nextRefreshDelay());
}

function animate() {
  requestAnimationFrame(animate);
  const time = performance.now() * 0.001;
  if (foodMesh) {
    foodMesh.rotation.y = time * 1.6;
    foodMesh.rotation.x = 0.45 + Math.sin(time * 1.9) * 0.08;
    foodMesh.position.y = 0.6 + Math.sin(time * 3.2) * 0.07;
    foodMesh.scale.setScalar(1 + Math.sin(time * 3.2) * 0.08);
  }
  foodMaterial.emissiveIntensity = 1.35 + (Math.sin(time * 3.2) + 1) * 0.18;
  foodGlowMaterial.opacity = 0.26 + (Math.sin(time * 3.2) + 1) * 0.05;
  snakeHeadMaterial.emissiveIntensity = 1.04 + (Math.sin(time * 4.4) + 1) * 0.07;
  snakeHeadGlowMaterial.opacity = 0.23 + (Math.sin(time * 4.4) + 1) * 0.03;
  if (foodGlowGroup.children[0]?.visible) {
    const glowPulse = 1.08 + Math.sin(time * 3.2) * 0.12;
    foodGlowGroup.children[0].scale.setScalar(glowPulse);
  }
  if (foodShadowGroup.children[0]?.visible) {
    const shadowPulse = 0.96 - Math.sin(time * 3.2) * 0.04;
    foodShadowGroup.children[0].scale.setScalar(shadowPulse);
  }
  if (snakeGlowGroup.children[0]?.visible) {
    const headPulse = 1.24 + Math.sin(time * 4.4) * 0.08;
    snakeGlowGroup.children[0].scale.setScalar(headPulse);
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
  const topOffset = Math.round((barRect?.bottom ?? 0) + layoutGap);
  const overlayVisible = !overlay.classList.contains("hidden");
  const overlayRect = overlayVisible ? overlay.getBoundingClientRect() : null;
  const overlayWidth = overlayRect ? Math.ceil(overlayRect.width) : 0;
  const overlayHeight = overlayRect ? Math.ceil(overlayRect.height) : 0;
  const sideBySide = overlayVisible && window.innerWidth >= narrowBreakpoint;

  let boardLeft = layoutGap;
  let boardTop = topOffset;
  let boardWidth = Math.max(0, window.innerWidth - layoutGap * 2);
  let boardHeight = Math.max(0, window.innerHeight - topOffset - layoutGap);

  if (overlayVisible) {
    if (sideBySide) {
      overlay.style.left = `${layoutGap}px`;
      overlay.style.top = `${topOffset}px`;

      boardLeft = layoutGap * 2 + overlayWidth;
      boardWidth = Math.max(0, window.innerWidth - boardLeft - layoutGap);
    } else {
      const overlayLeft = Math.max(layoutGap, Math.round((window.innerWidth - overlayWidth) / 2));
      const overlayTop = Math.max(topOffset + layoutGap, window.innerHeight - overlayHeight - layoutGap);
      overlay.style.left = `${overlayLeft}px`;
      overlay.style.top = `${overlayTop}px`;

      boardHeight = Math.max(0, overlayTop - topOffset - layoutGap);
    }
  }

  const size = Math.max(96, Math.floor(Math.min(boardWidth, boardHeight)));
  sceneViewport = {
    left: Math.round(boardLeft + Math.max(0, (boardWidth - size) / 2)),
    top: Math.round(boardTop + Math.max(0, (boardHeight - size) / 2)),
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

developerModeToggle.addEventListener("change", async (event) => {
  await updateDeveloperMode(event.target.checked);
});

developerLevelBtn.addEventListener("click", async () => {
  const level = Number.parseInt(developerLevelInput.value, 10);
  if (!Number.isFinite(level) || level < 1) {
    setStatus("Enter a level of 1 or greater", true);
    return;
  }
  await bypassLevel(level);
});

developerLevelInput.addEventListener("keydown", async (event) => {
  if (event.key !== "Enter") {
    return;
  }
  event.preventDefault();
  developerLevelBtn.click();
});

refresh();
updateViewport();
animate();
