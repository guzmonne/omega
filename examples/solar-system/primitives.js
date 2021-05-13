import * as THREE from "three"

export function createSphere({
  radius    = 1,
  wSegments = 6,
  hSegments = 6,
  color     = 0xFFFF00,
  geometry,
  material,
} = {}) {
  geometry || (geometry = new THREE.SphereGeometry(radius, wSegments, hSegments))
  material || (material = new THREE.MeshPhongMaterial({emissive: color}))
  return {geometry, material, mesh: new THREE.Mesh(geometry, material)}
}

export function createPointLight({
  color = 0xFFFFFF,
  intensity = 3,
} = {}) {
  return new THREE.PointLight(color, intensity)
}