import {makeRequest} from "./make-request";
import {SSMClient} from "@aws-sdk/client-ssm";
import {ParamStore, StravaParams} from "./param-store";

const config = {region: process.env.AWS_REGION};
const ssm = new SSMClient(config);
const paramStore = new ParamStore(ssm);

const http401 = {statusCode: 401, body: JSON.stringify("Unauthorized")};

export const handler = async (event: any) => {
    if (!event.queryStringParameters.code) {
        console.error("No 'code' parameter in query string parameters");
        return http401;
    }

    let params: StravaParams;
    try {
        params = await paramStore.getParams();
    } catch (e) {
        console.error(e);
        return http401;
    }

    const data = {
        client_id: params.clientId,
        client_secret: params.clientSecret,
        code: event.queryStringParameters.code,
        grant_type: "authorization_code",
    };
    const url = "https://www.strava.com/oauth/token?" + new URLSearchParams(data);

    let res: any;
    try {
        res = await makeRequest(url, {});
    } catch (e) {
        console.error("Get token failed", e);
        return http401;
    }

    await paramStore.setParam("refreshToken", res.refresh_token, "SecureString");
    await paramStore.setParam("accessToken", res.access_token, "SecureString");

    return {
        statusCode: 200,
        body: JSON.stringify("Authorized"),
    };
};
