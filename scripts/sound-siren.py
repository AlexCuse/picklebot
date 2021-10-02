import RPi.GPIO as GPIO
import time

sound = 12
blue = 17
red = 16
  
GPIO.setmode(GPIO.BCM)
GPIO.setup(sound, GPIO.IN)
GPIO.setup(blue, GPIO.OUT)
GPIO.setup(red, GPIO.OUT)
  
def callback(pin):
    if GPIO.input(pin):
        print("SOUND!")
        GPIO.output(blue, True)
        GPIO.output(red, False)

        time.sleep(.25)
     
        GPIO.output(red, True)
        GPIO.output(blue, False)
        
        time.sleep(.25)
        
        GPIO.output(red, False)
    else:
        print("quiet.....")
        
GPIO.add_event_detect(sound, GPIO.BOTH, bouncetime=500)
GPIO.add_event_callback(sound, callback)

while True:
    time.sleep(1)