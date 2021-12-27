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

const http401 = {statusCode: 401, body: JSON.stringify("Unauthorized")};

const ssm = new AWS.SSM();

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
        Name: param,
        Value: value,
        Type: type
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

exports.handler = async(event) => {
    if (!event.queryStringParameters.code) {
        console.error("No 'code' parameter in query string parameters");
        return http401;
    }

    let params;
    try {
        params = await getParams();
        console.log(params);
    } catch (e) {
        console.error(e);
        return http401;
    }

    const data = {
        client_id: params.clientId,
        client_secret: params.clientSecret,
        code: event.queryStringParameters.code,
        grant_type: "authorization_code"
    };
    const url = "https://www.strava.com/oauth/token?" + new URLSearchParams(data);

    let res;
    try {
        res = await makeRequest(url, {});
    } catch (e) {
        console.error("Get token failed", e);
        return http401;
    }

    if (res.athlete.id !== 21702111) {
        return http401;
    }

    await setParam("/strava/refreshToken", res.refresh_token, "SecureString");
    await setParam("/strava/accessToken", res.access_token, "SecureString");

    return {
        statusCode: 200, body: JSON.stringify("Authorized"),
    };
};
