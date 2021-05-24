import './style.css'
import * as THREE from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import * as dat from "dat.gui"

/**
 * Debug
 */
const gui = new dat.GUI()

/**
 * Textures
 */
const textureLoader     = new THREE.TextureLoader()
const cubeTextureLoader = new THREE.CubeTextureLoader()

const doorTexture = {
  color           : textureLoader.load("/assets/textures/door/color.jpg"),
  alpha           : textureLoader.load("/assets/textures/door/alpha.jpg"),
  ambientOcclusion: textureLoader.load("/assets/textures/door/ambientOcclusion.jpg"),
  height          : textureLoader.load("/assets/textures/door/height.jpg"),
  metalness       : textureLoader.load("/assets/textures/door/metalness.jpg"),
  normal          : textureLoader.load("/assets/textures/door/normal.jpg"),
  roughness       : textureLoader.load("/assets/textures/door/roughness.jpg"),
}
const matcapTexture   = textureLoader.load("/assets/matcaps/8.png")
const gradientTexture = textureLoader.load("/assets/gradients/3.jpg")
const environmentMap  = cubeTextureLoader.load([
  "/assets/environmentMaps/0/px.jpg",
  "/assets/environmentMaps/0/nx.jpg",
  "/assets/environmentMaps/0/py.jpg",
  "/assets/environmentMaps/0/ny.jpg",
  "/assets/environmentMaps/0/pz.jpg",
  "/assets/environmentMaps/0/nz.jpg",
])
gradientTexture.minFilter = THREE.NearestFilter
gradientTexture.magFilter = THREE.NearestFilter
gradientTexture.generateMipmaps = false
/**
 * Base
 */
// Canvas
const canvas = document.querySelector('canvas.webgl')

// Scene
const scene = new THREE.Scene()

/**
 * Objects
 */
// const material = new THREE.MeshBasicMaterial()
// const material = new THREE.MeshNormalMaterial()
// const material = new THREE.MeshMatcapMaterial()
// const material = new THREE.MeshLambertMaterial()
// const material = new THREE.MeshPhongMaterial()
// const material = new THREE.MeshToonMaterial()
const material = new THREE.MeshStandardMaterial()
// material.map             = doorTexture.color
// material.aoMap           = doorTexture.ambientOcclusion
// material.displacementMap = doorTexture.height
// material.metalnessMap    = doorTexture.metalness
// material.roughnessMap    = doorTexture.roughness
// material.normalMap       = doorTexture.normal
// material.alphaMap        = doorTexture.alpha
material.envMap = environmentMap

// material.displacementScale = 0.05
// material.transparent = true
material.metalness = 0.7
material.roughness = 0.2

gui.add(material, "metalness", 0, 1, .01)
gui.add(material, "roughness", 0, 1, .01)
gui.add(material, "aoMapIntensity", 0, 10, .0001)
gui.add(material, "displacementScale", 0, 10, .0001)
gui.add(material.normalScale, "x", 0, 1, .001)
gui.add(material.normalScale, "y", 0, 1, .001)

const sphere = new THREE.Mesh(new THREE.SphereBufferGeometry(0.5, 64, 64), material)
const plane  = new THREE.Mesh(new THREE.PlaneBufferGeometry(1, 1, 100, 100), material)
const torus  = new THREE.Mesh(new THREE.TorusBufferGeometry(0.3, 0.2, 64, 128), material)

sphere.geometry.setAttribute("uv2", new THREE.BufferAttribute(
  sphere.geometry.attributes.uv.array,
  2,
))
sphere.position.x = -1.1

plane.geometry.setAttribute("uv2", new THREE.BufferAttribute(
  plane.geometry.attributes.uv.array,
  2,
))

torus.geometry.setAttribute("uv2", new THREE.BufferAttribute(
  torus.geometry.attributes.uv.array,
  2,
))
torus.position.x = 1.1

scene.add(sphere, plane, torus)
/**
 * Lights
 */
const ambientLight = new THREE.AmbientLight(0xffffff, 0.5)
const pointLight   = new THREE.PointLight(0xffffff, 0.5)

pointLight.position.x = 2
pointLight.position.y = 3
pointLight.position.z = 4

scene.add(ambientLight, pointLight)
/**
 * Sizes
 */
const sizes = {
  width: window.innerWidth,
  height: window.innerHeight
}

window.addEventListener('resize', () => {
  // Update sizes
  sizes.width = window.innerWidth
  sizes.height = window.innerHeight

  // Update camera
  camera.aspect = sizes.width / sizes.height
  camera.updateProjectionMatrix()

  // Update renderer
  renderer.setSize(sizes.width, sizes.height)
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
})

/**
 * Camera
 */
// Base camera
const camera = new THREE.PerspectiveCamera(75, sizes.width / sizes.height, 0.1, 100)
camera.position.x = 1
camera.position.y = 1
camera.position.z = 2
scene.add(camera)

// Controls
const controls = new OrbitControls(camera, canvas)
controls.enableDamping = true

/**
 * Renderer
 */
const renderer = new THREE.WebGLRenderer({
    canvas: canvas
})
renderer.setSize(sizes.width, sizes.height)
renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))

/**
 * Animate
 */
const clock = new THREE.Clock()

const tick = () => {
  const elapsedTime = clock.getElapsedTime()

  sphere.rotation.x = 0.1 * elapsedTime
  plane.rotation.x  = 0.1 * elapsedTime
  torus.rotation.x  = 0.1 * elapsedTime

  sphere.rotation.y = 0.1 * elapsedTime
  plane.rotation.y  = 0.1 * elapsedTime
  torus.rotation.y  = 0.1 * elapsedTime

  // Update controls
  controls.update()

  // Render
  renderer.render(scene, camera)

  // Call tick again on the next frame
  window.requestAnimationFrame(tick)
}

tick()