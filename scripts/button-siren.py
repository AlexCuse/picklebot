import RPi.GPIO as GPIO
import time

try:

  GPIO.setmode(GPIO.BCM)

  button = 19
  GPIO.setup(button, GPIO.IN, pull_up_down=GPIO.PUD_UP)
  
  blue = 17
  red = 16
  GPIO.setup(blue, GPIO.OUT)
  GPIO.setup(red, GPIO.OUT)

  while True:
    input_state = GPIO.input(button)
    
    if input_state == False:
        for x in range(0,10):
            GPIO.output(blue, True)
            GPIO.output(red, False)

            time.sleep(.25)
         
            GPIO.output(red, True)
            GPIO.output(blue, False)
            
            time.sleep(.25)
    else: 
      GPIO.output(blue, False)
      GPIO.output(red, False)

except KeyboardInterrupt:
  print("exiting")

finally:
  GPIO.cleanup()