const https = require("https");

export const makeRequest = (path: string, payload: any, method = "POST", headers = {}) =>
    new Promise((resolve, reject) => {
        const options = {host: "strava.com", port: 443, path, method, headers};
        const req = https.request(options, (res: any) => {
            let buffer = "";
            res.on("data", (chunk: string) => (buffer += chunk));
            res.on("end", () => {
                resolve(JSON.parse(buffer));
            });
        });
        req.on("error", (e: any) => reject(e.message));
        req.write(JSON.stringify(payload));
        req.end();
    });
