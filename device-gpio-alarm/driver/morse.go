package driver

import (
	"time"

	"github.com/warthog618/gpiod"
)

/*
	We will use morse code to send a 'message' through the configured alarm pin (in example an LED)

	International Morse code is composed of five elements:

	    - short mark, dot or dit ('.'): "dit duration" is one time unit long.
	    - long mark, dash or dah ('-'): three time units long.
	    - inter-element gap between the dits and dahs within a character: one dot duration or one unit long.
	    - short gap (' '): three time units long.
	    - medium gap ('/'): seven time units long.
*/

var (
	morseUnit = 50 * time.Millisecond

	morse = map[rune]time.Duration{
		'.': 1 * morseUnit,
		'-': 3 * morseUnit,
		'/': 7 * morseUnit,
		' ': 3 * morseUnit,
	}
)

func sendMorse(msg string, ap int) error {
	line, err := gpiod.RequestLine(chip, ap, gpiod.AsOutput(0))

	if err != nil {
		return err
	}

	defer func() {
		line.Reconfigure(gpiod.AsInput)
		line.Close()
	}()

	for _, c := range msg {
		blip := time.Duration(0)

		switch c {
		case '.', '-':
			line.SetValue(1)
			// this creates the tiny hesitation between elements of a character
			blip = morseUnit
		}

		time.Sleep(morse[c])

		line.SetValue(0)
		time.Sleep(blip)
	}
	return nil
}
