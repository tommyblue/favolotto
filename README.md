# Favolotto

![Favolotto](favolotto.png)

## Hardware

- [Raspberry Pi Zero 2 W](https://www.raspberrypi.com/products/raspberry-pi-zero-2-w/) or [Raspberry Pi Zero W](https://www.raspberrypi.com/products/raspberry-pi-zero-w/)
- [Keyestudio 5V ReSpeaker 2-Mic Pi HAT V1.0](https://www.keyestudio.com/products/keyestudio-5v-respeaker-2-mic-pi-hat-v10-expansion-board-for-raspberry-pi-3b-4b)
- PN532 NFC reader
- 3 Watt speaker
- MicroSD card
- Power bank
- Adehesive NFC tags (Ntag 215)
- 3D printed case

## Software

Use [Raspberry Pi Imager](https://www.raspberrypi.com/software/) to prepare the MicroSD card with the Raspberry Pi OS Lite 64bit, using custom setup to configure wireless settings. Also, enable SSH.
Boot the Rpi with the SD card inserted and connecto to it through SSH.
Run `sudo raspi-config`and enable I2C (`Interface Options > I2C`). Don't forget to expand the filesystem from `Advanced Options > Expand Filesystem`.

### Audio hat

Install the I2C package and check the sound hat is correctly detected (the detected value in the matrix could be different):

```sh
sudo apt install -y i2c-tools git
sudo i2cdetect -y 1

     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
00:                         -- -- -- -- -- -- -- --
10: -- -- -- -- -- -- -- -- -- -- 1a -- -- -- -- --
20: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
30: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
40: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
50: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
60: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
70: -- -- -- -- -- -- -- --
```

Install the hat drivers:

```sh
git clone https://github.com/waveshare/WM8960-Audio-HAT
cd WM8960-Audio-HAT/
sudo ./install.sh
```

Edit the configuration with `sudo vi /boot/firmware/config.txt` and add:

```
dtparam=audio=off # this is already present with "on", change it
dtoverlay=wm8960-soundcard
```

After `reboot`, you shoud see the sound card:

```sh
$ aplay -l

**** List of PLAYBACK Hardware Devices ****
card 0: wm8960soundcard [wm8960-soundcard], device 0: 3f203000.i2s-wm8960-hifi wm8960-hifi-0 [3f203000.i2s-wm8960-hifi wm8960-hifi-0]
  Subdevices: 1/1
  Subdevice #0: subdevice #0
card 1: vc4hdmi [vc4-hdmi], device 0: MAI PCM i2s-hifi-0 [MAI PCM i2s-hifi-0]
  Subdevices: 1/1
  Subdevice #0: subdevice #0
```

Copy an mp3 file to the Rpi and check it works with `mpg321 file.mp3` (`sudo apt install mpg321` if it is not available). If you don't hear anything, check the volume with `alsamixer` (use `m` to mute/unmute channels).

### NFC (WIP)

```sh
sudo apt install libnfc6 libnfc-bin libnfc-examples
nfc-list
nfc-poll
```
#### NFC PN7150
Using the shield with the PN7150 here you could find the reference guide:

https://community.nxp.com/t5/NXP-Designs-Knowledge-Base/Easy-  git clone https://github.com/NXPNFCLinux/linux_libnfc-nci.gitset-up-of-NFC-on-Raspberry-Pi/ta-p/1099034

Here you could find the main example:

git clone https://github.com/NXPNFCLinux/linux_libnfc-nci.git

*Note*: demo and lib needs to have connected the Int pin.

#### NFC PN7150 Connection:

![immagine](https://github.com/user-attachments/assets/16267140-17fc-418d-b906-ea131f02a79d)

- Signal | Pin No.
- SDA -> GPIO 2
- SCL -> GPIO 3
- Vdd -> 3.3V
- GND -> GND
- Int -> GPIO23

### Result

This is an initial prototype:

![Prototype](prototype.jpg)
