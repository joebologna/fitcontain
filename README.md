# Overlaying a fyne.CanvasObject on a Background Image

When using a background image, such at the playing field of a game, it is necessary to use "fit contain" to keep the image oriented properly, centered and use all available space. 

In this diagnostic program the behavior of the algorithm to keep the fyne.CanvasObject in the same relative position over the image is demonstrated by:

- showing the red circle location in the global space. 
- showing the teal circle overlaid on the sample image.
- as the window is resized, the teal circle will be in the same position as the red circle.
- when the aspect ratio of the global space matches the aspect ratio of the sample image.
- both circles will stay in the middle of the upper left quadrant of the space it occupies.
- as the size of image changes, the teal circle changes size accordingly.

This technique can be used to place a background image and overlay buttons that stay in the correct location regardless of window size.

This is useful for portability across desktop and mobile devices.

![](./demo.gif)
