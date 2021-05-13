import {renderer, camera, scene} from "./main"

export function resizeRenderer(renderer) {
  const canvas = renderer.domElement
  const width  = canvas.clientWidth
  const height = canvas.clientHeight

  const needResize = window.innerWidth !== width || window.innerHeight !== height

  if (needResize) {
    renderer.setSize(window.innerWidth, window.innerHeight, false)
  }

  console.log(needResize)

  return needResize
}

export function createAnimateFn(callback) {
  return function animate(time) {
    time *= 0.001

    if (resizeRenderer(renderer)) {
      const canvas = renderer.domElement
      camera.aspect = canvas.clientWidth / canvas.clientHeight
      camera.updateProjectionMatrix()
    }

    callback(time)

    renderer.render(scene, camera)

    requestAnimationFrame(animate)
  }
}