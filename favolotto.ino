#include <Wire.h>
#include <Adafruit_GFX.h>
#include <Adafruit_ILI9341.h>
#include <Adafruit_PN532.h>
// #include <DFRobotDFPlayerMini.h>
#include <Arduino.h>
#include <WiFi.h>
#include <ESPAsyncWebServer.h>
#include <SD.h>
#include <SPI.h>
#include "SPIFFS.h"

#include "credentials.h"

extern const char *SSID;
extern const char *PASSWORD;

// Configurazioni hardware
#define TFT_CS 5
#define TFT_RST 22
#define TFT_DC 21
#define SDA_PIN 21  // Pin SDA per I2C
#define SCL_PIN 22  // Pin SCL per I2C
// #define RX_PIN 16   // Pin RX per DFPlayer Mini
// #define TX_PIN 17   // Pin TX per DFPlayer Mini

// Definizione pin per SPI SD
#define SD_CS_PIN 5  // Chip Select per microSD
#define SD_MOSI 23
#define SD_MISO 19
#define SD_SCK 18

// Inizializza il display TFT
// Adafruit_ILI9341 tft = Adafruit_ILI9341(TFT_CS, TFT_DC, TFT_RST);

// Inizializza il lettore NFC
Adafruit_PN532 nfc(SDA_PIN, SCL_PIN);

// Inizializza il DFPlayer per la riproduzione MP3
// DFRobotDFPlayerMini myDFPlayer;

// Inizializzazione del server
// AsyncWebServer server(80);


// const char INDEX_HTML[] PROGMEM = R"=====(
// <!DOCTYPE html>
// <html lang="en">
//   <head>
//     <meta charset="UTF-8" />
//     <meta name="viewport" content="width=device-width, initial-scale=1.0" />
//     <title>Vite + React</title>
//     <script type="module" crossorigin src="/index.js"></script>
//     <link rel="stylesheet" crossorigin href="/index.css">
//   </head>
//   <body>
//     <div id="root"></div>
//   </body>
// </html>
// )=====";

// const char INDEX_JS[] PROGMEM = R"=====(
// (alert("here"))();
// )=====";

// const char INDEX_CSS[] PROGMEM = R"=====(
// body {background-color:red;}
// )=====";


void setup() {
  Serial.begin(115200);
  Serial.println("Inizializzazione del sistema...");

  // Connessione WiFi
  WiFi.begin(SSID, PASSWORD);
  while (WiFi.status() != WL_CONNECTED) {
    delay(1000);
    Serial.println("Connessione WiFi...");
  }
  Serial.println("Connesso!");
  IPAddress myIP = WiFi.localIP();
  Serial.println(myIP);

  // if (!SPIFFS.begin(true)) {
  //   Serial.println("Errore nell'inizializzazione di SPIFFS!");
  //   return;
  // }

  // Serial.println("SPIFFS montato correttamente!");

  // File fileHtml = SPIFFS.open("/index.html", "w");
  // if (!fileHtml) {
  //   Serial.println("Errore nella creazione del file");
  //   return;
  // }
  // fileHtml.println(INDEX_HTML);
  // fileHtml.close();

  // File fileJs = SPIFFS.open("/index.js", "w");
  // if (!fileJs) {
  //   Serial.println("Errore nella creazione del file");
  //   return;
  // }
  // fileJs.println(INDEX_JS);
  // fileJs.close();

  // File fileCss = SPIFFS.open("/index.css", "w");
  // if (!fileCss) {
  //   Serial.println("Errore nella creazione del file");
  //   return;
  // }
  // fileCss.println(INDEX_CSS);
  // fileCss.close();


  // Inizializzazione SD
  SPI.begin(SD_SCK, SD_MISO, SD_MOSI, SD_CS_PIN);
  if (!SD.begin(SD_CS_PIN)) {
    Serial.println("Errore nell'inizializzazione della SD!");
    // return;
  }

  Serial.println("SD pronta!");

  listSDFiles();

  // Avvia il display TFT
  // tft.begin();
  // tft.setRotation(3);
  // tft.fillScreen(ILI9341_WHITE);

  // Avvia il lettore NFC
  nfc.begin();
  uint32_t versiondata = nfc.getFirmwareVersion();
  if (!versiondata) {
    Serial.println("Impossibile trovare il lettore NFC");
  } else {
    Serial.println(versiondata);
  }

  // Avvia il DFPlayer utilizzando Serial2
  // Serial2.begin(9600, SERIAL_8N1, RX_PIN, TX_PIN);
  // if (!myDFPlayer.begin(Serial2)) {
  //   Serial.println("DFPlayer non trovato");
  // }
  // myDFPlayer.volume(20);
  // myDFPlayer.play(1);

  // int value = myDFPlayer.readFileCounts();
  // if (value == -1) {  //Error while Reading.
  //   printDetail(myDFPlayer.readType(), myDFPlayer.read());
  // } else {  //Successfully get the result.
  //   Serial.println(value);
  // }

  // if (myDFPlayer.available()) {
  //   printDetail(myDFPlayer.readType(), myDFPlayer.read());  //Print the detail message from DFPlayer to handle different errors and states.
  // }

  // Endpoint GET per ottenere la lista delle storie
  // server.on("/stories", HTTP_GET, [](AsyncWebServerRequest *request) {
  //   File root = SD.open("/");
  //   if (!root) {
  //     request->send(500, "text/plain", "Errore apertura SD");
  //     return;
  //   }

  //   String json = "[";
  //   File file = root.openNextFile();
  //   while (file) {
  //     if (!file.isDirectory() && String(file.name()).endsWith(".mp3")) {
  //       json += "{\"name\":\"" + String(file.name()) + "\"}";
  //       if (root.openNextFile()) json += ",";
  //     }
  //     file.close();
  //     file = root.openNextFile();
  //   }
  //   json += "]";
  //   request->send(200, "application/json", json);
  // });

  // // Endpoint POST per caricare una storia (MP3 e immagine)
  // server.on("/stories", HTTP_POST, [](AsyncWebServerRequest *request) {
  //   request->send(200, "text/plain", "Caricamento non ancora implementato");
  // });

  // // Servire l'app React dalla SD
  // server.on("/", HTTP_GET, [](AsyncWebServerRequest *request) {
  //   // request->send(SD, "/index.html", "text/html");
  //   request->send(SPIFFS, "/index.html", "text/html;charset=UTF-8");
  // });

  // server.on("/index.js", HTTP_GET, [](AsyncWebServerRequest *request) {
  //   request->send(SPIFFS, "/index.js", "text/javascript;charset=UTF-8");
  // });

  // server.on("/index.css", HTTP_GET, [](AsyncWebServerRequest *request) {
  //   request->send(SPIFFS, "/index.css", "text/css;charset=UTF-8");
  // });

  // server.begin();

  Serial.println("Sistema pronto!");
}

void loop() {
  uint8_t success;
  uint8_t uid[] = { 0, 0, 0, 0, 0, 0, 0 };
  uint8_t uidLength;
  success = nfc.readPassiveTargetID(PN532_MIFARE_ISO14443A, uid, &uidLength);

  if (success) {
    Serial.print("Tag NFC rilevato: ");
    for (int i = 0; i < uidLength; i++) {
      Serial.print(" 0x");
      Serial.print(uid[i], HEX);
    }
    Serial.println("");
  }
}

// void printDetail(uint8_t type, int value) {
//   switch (type) {
//     case TimeOut:
//       Serial.println(F("Time Out!"));
//       break;
//     case WrongStack:
//       Serial.println(F("Stack Wrong!"));
//       break;
//     case DFPlayerCardInserted:
//       Serial.println(F("Card Inserted!"));
//       break;
//     case DFPlayerCardRemoved:
//       Serial.println(F("Card Removed!"));
//       break;
//     case DFPlayerCardOnline:
//       Serial.println(F("Card Online!"));
//       break;
//     case DFPlayerUSBInserted:
//       Serial.println("USB Inserted!");
//       break;
//     case DFPlayerUSBRemoved:
//       Serial.println("USB Removed!");
//       break;
//     case DFPlayerPlayFinished:
//       Serial.print(F("Number:"));
//       Serial.print(value);
//       Serial.println(F(" Play Finished!"));
//       break;
//     case DFPlayerError:
//       Serial.print(F("DFPlayerError:"));
//       switch (value) {
//         case Busy:
//           Serial.println(F("Card not found"));
//           break;
//         case Sleeping:
//           Serial.println(F("Sleeping"));
//           break;
//         case SerialWrongStack:
//           Serial.println(F("Get Wrong Stack"));
//           break;
//         case CheckSumNotMatch:
//           Serial.println(F("Check Sum Not Match"));
//           break;
//         case FileIndexOut:
//           Serial.println(F("File Index Out of Bound"));
//           break;
//         case FileMismatch:
//           Serial.println(F("Cannot Find File"));
//           break;
//         case Advertise:
//           Serial.println(F("In Advertise"));
//           break;
//         default:
//           break;
//       }
//       break;
//     default:
//       break;
//   }
// }

void listSDFiles() {
  File root = SD.open("/");
  if (!root) {
    Serial.println("Errore apertura SD!");
    return;
  }

  Serial.println("File nella microSD:");
  File file = root.openNextFile();
  while (file) {
    Serial.println(file.name());
    file.close();
    file = root.openNextFile();
  }
}