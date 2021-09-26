import RPi.GPIO as GPIO
import time
 
blue = 17 # BOARD: 11
red = 16
 
GPIO.setmode(GPIO.BCM)
 
GPIO.setup(blue, GPIO.OUT)
GPIO.setup(red, GPIO.OUT)

for x in range(0,10):
    GPIO.output(blue, True)
    GPIO.output(red, False)

    time.sleep(.25)
 
    GPIO.output(red, True)
    GPIO.output(blue, False)
    
    time.sleep(.25)
    
GPIO.cleanup()
