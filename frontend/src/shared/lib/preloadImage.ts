export function preloadImage(source: string): Promise<void> {
  return new Promise<void>((resolve) => {
    const image = new Image()
    image.src = source

    const finish = () => resolve()
    image.addEventListener('load', finish, { once: true })
    image.addEventListener('error', finish, { once: true })

    if (image.complete) finish()
    else if (typeof image.decode === 'function')
      image
        .decode()
        .then(finish)
        .catch(() => undefined)
  })
}
