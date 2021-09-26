import RPi.GPIO as GPIO
import time
 
blue = 17 # BOARD: 11
red = 16
 
GPIO.setmode(GPIO.BCM)
 
GPIO.setup(blue, GPIO.OUT)
GPIO.setup(red, GPIO.OUT)

GPIO.output(blue, True)
GPIO.output(red, True)

time.sleep(5)
 
GPIO.output(red, False)
GPIO.output(blue, False)