const https = require("https");
const AWS = require("aws-sdk")

const makeRequest = (path, payload, method = "POST", headers = {}) => new Promise((resolve, reject) => {
    const options = {host: "strava.com", port: 443, path, method, headers};
    const req = https.request(options, res => {
        let buffer = "";
        res.on("data", chunk => buffer += chunk);
        res.on("end", () => {
            resolve(JSON.parse(buffer))
        });
    });
    req.on("error", e => reject(e.message));
    req.write(JSON.stringify(payload));
    req.end();
});

const ssm = new AWS.SSM();
const sns = new AWS.SNS();

const getParams = () => {
    return new Promise((r, x) => {
        const params = {
            Path: "/strava", Recursive: true, WithDecryption: true
        };
        ssm.getParametersByPath(params, function(err, data) {
            if (err) {
                x(err);
            } else {
                const map = data.Parameters.reduce((acc, val) => {
                    const key = val.Name.replace("/strava/", "");
                    acc[key] = val.Value;
                    return acc;
                }, {});
                r(map);
            }
        });
    });
};

const setParam = (param, value, type = "String") => {
    const data = {
        Name: param, Value: value, Type: type, Overwrite: true
    };

    return new Promise((r, x) => {
        ssm.putParameter(data, function(err, data) {
            if (err) {
                x(err);
            } else {
                r(data);
            }
        });
    });
};

const publishSns = (params) => {
    return new Promise((r,x) => {
        sns.publish(params, function(err, data) {
            if (err) {
                x(err);
            } else {
                r(data);
            }
        });
    });
};

exports.handler = async(event) => {

    let ssmParams;
    try {
        ssmParams = await getParams();
    } catch (e) {
        console.error(e);
        return false;
    }

    // //Refresh access token
    const data = {
        client_id: ssmParams.clientId,
        client_secret: ssmParams.clientSecret,
        grant_type: "refresh_token",
        refresh_token: ssmParams.refreshToken,
    };
    const url = "https://www.strava.com/oauth/token?" + new URLSearchParams(data);

    let res;
    try {
        res = await makeRequest(url, {});
    } catch (e) {
        console.error("Get token failed");
        return false;
    }

    await setParam("/strava/refreshToken", res.refresh_token, "SecureString");
    await setParam("/strava/accessToken", res.access_token, "SecureString");

    ssmParams.refreshToken = res.refresh_token;
    ssmParams.accessToken = res.access_token;

    //Load activities
    try {
        const headers = {Authorization: `Bearer ${ssmParams.accessToken}`};
        res = await makeRequest("https://www.strava.com/api/v3/athlete/activities", {}, "GET", headers);
    } catch (e) {
        return false;
    }

    const types = ["Run", "Hike", "Walk"];
    const badShoes = ["g9558316"];

    const needFixing = res.filter((act) => {
        if (!types.includes(act.type)) {
            return false;
        }
        return !act.gear_id || badShoes.includes(act.gear_id);
    });

    needFixing.forEach((act) => {
        console.log(act.name, act.type, act.gear_id);
    });

    if (needFixing.length) {
        const Message = needFixing.map((a) => `${a.name} (${a.type}) https://www.strava.com/activities/${a.id}`).join("\n");

        const params = {
            Message,
            Subject: "Strava activities with missing footwear",
            TopicArn: ssmParams.topicId
        };
        await publishSns(params);
    } else {
        console.info("No missing footwear");
    }
};
