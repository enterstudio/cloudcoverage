# Notes

## How it currently works

### On Raspberry Pi

- Scheduled [node script](https://github.com/ubergesundheit/senseBox-cloud/blob/7040bc09c37e4beb707732450eb46c7647a951ae/raspberrypi/index.js)
  - Takes lightness reading of TSL45315 sensor through external [C code](https://github.com/ubergesundheit/senseBox-cloud/blob/7040bc09c37e4beb707732450eb46c7647a951ae/raspberrypi/tsl.c) which just returns a float.
  - Calculates required shutter speed for `raspistill` with lighness reading
    ```javascript
    // ignores decimals
    var lux = parseFloat(data.toString('utf-8').split('.')[0])
    var shutterSpeed = (0.7368 * Math.pow(lux.toFixed(6), -0.915)) * 1000000
    ```
  - Invokes `raspistill -o image.jpg -t 1 -ss shutterSpeed --metering backlit --ISO 100 --awb sun --nopreview`
    - If luminance is 0, shutterSpeed is set to 6000
  - Reads the saved `image.jpg` and sends it to the AWS Lambda function as base64 encoded string together with timestamp, senseBox Id and sensor Id
  
  
### On AWS Lambda function

- On HTTP Post:
  - Image is decoded from base64 string
  - Image is flipped horizontally (unnecessary?)
  - Lens barrel distortion (imagemagick `convert -distort barrel 0.005 -0.025 -0.028`)
  - Image is flipped horizontally (unnecessary?)
  - Image is rotated by 180 degrees clockwise
  - Image is flipped horizontally
  - Location of senseBox Id is requested from opensensemap-api
  - Sunposition is calculated using [SunCalc](https://github.com/mourner/suncalc) with the timestamp from the HTTP POST and the location of the senseBox from the API response
      ```javascript
      // current sun position
      var sunPos = SunCalc.getPosition(new Date(req.body.timestamp), location.lat, location.lon)
      // convert to degrees
      sunPos.azimuth *= (180 / Math.PI)
      sunPos.azimuth += 180

      // factor 0.7 for a big kegel
      sunPos.altitude = sunPos.altitude * (180 / Math.PI) * 0.7;

      // center of the image
      var centerx = Math.floor(image.bitmap.width / 2)
      var centery = Math.floor(image.bitmap.height / 2)

      // min and max for the kegel
      var azimin = ((sunPos.azimuth - sunPos.altitude) < 0 ? 360 - Math.abs(sunPos.azimuth - sunPos.altitude) : sunPos.azimuth - sunPos.altitude);
      var azimax = (sunPos.azimuth + sunPos.altitude) > 360 ? 360 - sunPos.azimuth + sunPos.altitude : sunPos.azimuth + sunPos.altitude;
      ```
  - Each pixel is then iterated and the following calculation is done:
    - Angle between current pixel and center is calculated
    - If the angle is bigger than azimin and the angle is less than azimax and sunPos.altitude is bigger than 15
      - Count the pixel as sun pixel
    - Otherwise the color of the pixel is taken
      - Red channel is divided by the blue channel (`rbr = r / b`)
      - If all channels (red, green, blue) are above 250, the pixel is also the sun
      - The pixel is a cloud, if the `rbr` value is bigger or equal to 0.85 and is counted
  - Cloud coverage is then `((cloudCounter / ((image.bitmap.width * image.bitmap.height)-sunCounter)) * 100)`
  - Cloud coverage of 4 images is averaged and then HTTP Post ed to openSenseMap

 
 ### Notes on the Algorithm
 - raspistill can do flipping
 - All the flipping and rotating can be done with a single vertical flip. Consider barrel distortion though!
 - Also is the flipping required at all? Sun azimuth angle is also added a 180, so is it possible to not rotate/flip and leave the 180 from the sun azimuth?
 
 
 
