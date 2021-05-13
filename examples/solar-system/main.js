import * as THREE from "three"
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';


var renderer, camera, scene

export function main() {
  renderer = new THREE.WebGLRenderer()
  renderer.setSize(window.innerWidth, window.innerHeight);
  document.body.appendChild(renderer.domElement);

  const fov    = 40
  const aspect = 2
  const near   = 0.1
  const far    = 1000

  camera = new THREE.PerspectiveCamera(fov, aspect, near, far)

  camera.position.set(0, 50, 0)
  camera.up.set(0, 0, 1)
  camera.lookAt(0, 0, 0)

  scene = new THREE.Scene()
  const controls = new OrbitControls(camera, renderer.domElement);
}

export {renderer, camera, scene}