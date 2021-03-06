var Jimp = require('jimp')
var request = require('request')
var SunCalc = require('suncalc')
var fs = require('fs')
var im = require('imagemagick');
var AWS = require('aws-sdk');
var s3 = new AWS.S3();

var averageCoverage = []

var s3BucketName = ''

exports.handler = (event, context, callback) => {

    var req = {};
    req.body = event;

    /**
    // http://stackoverflow.com/a/26815894
    var dir = './images';
    if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir);
    }
    */

    var tempImagePath = '/tmp/image_decoded.jpg';

    // http://stackoverflow.com/a/8111863
    var decodedImage = new Buffer(req.body.image, 'base64');
    fs.writeFileSync(tempImagePath, decodedImage);

    // flip image
    Jimp.read(tempImagePath, function(err, image) {

      image.flip(true, false)
      image.write(tempImagePath)
      image.flip(true, false)

      fs.readFile(tempImagePath, function(err, data) {
          s3.putObject({
              Bucket: s3BucketName,
              Key: "image_" + req.body.timestamp + ".jpg",
              Body: data,
              ContentType: 'JPG'
          }, function(err, data) {
              if (err) console.log(err, err.stack); // an error occurred
              else console.log(data); // successful response
          });
      })
    })

    // lens barrel distortion
    im.convert([
            tempImagePath,
            '-distort',
            'barrel',
            '0.005 -0.025 -0.028',
            tempImagePath
        ],
        function(err, stdout) {
            if (err) throw err;

            Jimp.read(tempImagePath, function(err, image) {

              image.flip(true, false)
              image.write(tempImagePath)
              image.flip(true, false)


                // counts the cloud pixels
                var counter = 0;

                // required, otherwise the sun is calculated on the wrong side
                image.rotate(180)
                image.flip(true, false)

                // dummy location near Muenster
                var location = {
                    lon: 7.6,
                    lat: 51.9
                }

                // GET location of sensesBox trough senseBox API
                request('https://api.opensensemap.org/boxes/' + req.body.sensebox_id, function(error, response, body) {
                    if (!error && response.statusCode == 200) {
                        console.log(JSON.parse(body.toString()).loc[0].geometry.coordinates)
                        location.lon = (JSON.parse(body.toString())).loc[0].geometry.coordinates[0]
                        location.lat = (JSON.parse(body.toString())).loc[0].geometry.coordinates[1]

                        // current sun position
                        var sunPos = SunCalc.getPosition(new Date(req.body.timestamp), location.lat, location.lon)
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

                        console.log(sunPos)
                        console.log({
                            azimin,
                            azimax
                        })

                        // double loop for each pixel in image
                        for (var x = 0; x < image.bitmap.width; x++) {
                            for (var y = 0; y < image.bitmap.height; y++) {

                                // calculate angle of current pixel
                                var dx = Math.abs(x - centerx);
                                var dy = Math.abs(y - centery);

                                // angle between center and current pixel
                                var alpha = 0

                                if (x > centerx) {
                                    if (y > centery) {
                                        alpha = Math.atan(dx / dy) * (180 / Math.PI);
                                    } else {
                                        alpha = (Math.atan(dy / dx) * (180 / Math.PI)) + 90;
                                    }
                                } else {
                                    if (y > centery) {
                                        alpha = (Math.atan(dy / dx) * (180 / Math.PI)) + 270;
                                    } else {
                                        alpha = (Math.atan(dx / dy) * (180 / Math.PI)) + 180;
                                    }
                                }
                                // TODO: add thereshold when sun has no impact on image
                                if (alpha > azimin && alpha < azimax && sunPos.altitude > 15) {
                                    image.setPixelColor(0x000000ff, x, y);
                                } else {

                                    var hex = image.getPixelColor(x, y)
                                    var rgb = Jimp.intToRGBA(hex)
                                    var rbr = rgb.r / rgb.b
                                    if (rgb.r > 250 && rgb.g > 250 && rgb.b > 250) { // volle sättigung -> sonne
                                        image.setPixelColor(0xffffffff, x, y); // white
                                    } else {
                                        counter++
                                        if (rbr >= 1) {
                                            image.setPixelColor(0xff0000ff, x, y); // rot
                                        } else if (rbr < 1 && rbr >= 0.95) {
                                            image.setPixelColor(0xfffc00ff, x, y); // gelb
                                        } else if (rbr < 0.95 && rbr >= 0.85) {
                                            image.setPixelColor(0x1cff00ff, x, y); // gruen
                                        } else {
                                            image.setPixelColor(0x400ffff, x, y); // blau
                                            counter--
                                        }
                                    }
                                }
                            }
                        }
                        // calculate coverage in %
                        var coverage = ((counter / (image.bitmap.width * image.bitmap.height)) * 100)
                        console.log('coverage: ' + coverage)

                        // TODO: currently only one box is supported
                        averageCoverage.push(coverage)

                        console.log('averageCoverage.length ' + averageCoverage.length);

                        // use 6 measurements to calculate the average coverage during the last interval
                        if (averageCoverage.length === 6) {
                            var sum = 0
                            for (var i = 0; i < averageCoverage.length; i++) {
                                sum += averageCoverage[i]
                            }
                            var avg = sum / averageCoverage.length

                            // post data to opensensemap
                            postToOSeM(avg, req.body.sensebox_id, req.body.sensor_id, new Date(req.body.timestamp))

                            averageCoverage = [] // flush array

                            callback(null, "success");
                        }



                        image.rotate(180)
                        image.flip(true, false)
                        image.write('/tmp/image_classified.jpg')

                        fs.readFile('/tmp/image_classified.jpg', function(err, data) {
                            s3.putObject({
                                Bucket: s3BucketName,
                                Key: "image_" + req.body.timestamp + "_classified.jpg",
                                Body: data,
                                ContentType: 'JPG'
                            }, function(err, data) {
                                if (err) console.log(err, err.stack); // an error occurred
                                else console.log(data); // successful response
                            });
                        })

                    }
                })

            });
        });
}

/*
 *  post coverage to openSenseMap
 *  @param coverage the measured coverage
 */
function postToOSeM(coverage, sensebox_id, sensor_id, timestamp) {
    // post measured coverage
    request.post({
        url: 'https://api.opensensemap.org/boxes/' + sensebox_id + '/' + sensor_id,
        form: {
            value: coverage.toFixed(2).toString(),
            createdAt: timestamp.toISOString()
        }
    }, function(err, httpResponse, body) {
        console.log('own post respond: ' + httpResponse.statusCode)
    });

    // get ifgi ceilometer coverage and post it
    request('http://www.uni-muenster.de/Klima/data/0017behg_de.txt', function(error, response, body) {
        if (!error && response.statusCode == 200) {
            var ifgiCoverage = body.split('/8')[0]
            request.post({
                url: 'https://api.opensensemap.org/boxes/' + sensebox_id + '/5784ec9a6078ab1200a4f73d',
                form: {
                    value: ifgiCoverage,
                    createdAt: (new Date()).toISOString()
                }
            }, function(err, httpResponse, body) {
                console.log('ifgi post respond: ' + httpResponse.statusCode)
            });
        }
    })
}
