import {makeRequest} from "./make-request";
import {SSMClient} from "@aws-sdk/client-ssm";
import {ParamStore, StravaParams} from "./param-store";
import {PublishCommand, PublishCommandInput, SNSClient} from "@aws-sdk/client-sns";

const config = {region: process.env.AWS_REGION};
const ssm = new SSMClient(config);
const paramStore = new ParamStore(ssm);
const sns = new SNSClient(config);

const publishSns = (params: PublishCommandInput) => {
    return sns.send(new PublishCommand(params));
};

export const handler = async (): Promise<void> => {
    let ssmParams: StravaParams;
    try {
        ssmParams = await paramStore.getParams();
    } catch (e) {
        console.error(e);
        return;
    }

    // //Refresh access token
    const data = {
        client_id: ssmParams.clientId,
        client_secret: ssmParams.clientSecret,
        grant_type: "refresh_token",
        refresh_token: ssmParams.refreshToken,
    };
    const url = "https://www.strava.com/oauth/token?" + new URLSearchParams(data);

    let refreshTokenRes: any;
    try {
        refreshTokenRes = await makeRequest(url, {});
    } catch (e) {
        console.error("Get token failed");
        return;
    }

    await paramStore.setParam("refreshToken", refreshTokenRes.refresh_token, "SecureString");
    await paramStore.setParam("accessToken", refreshTokenRes.access_token, "SecureString");

    ssmParams.refreshToken = refreshTokenRes.refresh_token;
    ssmParams.accessToken = refreshTokenRes.access_token;

    //Load activities
    let activitiesRes: any[];
    try {
        const headers = {Authorization: `Bearer ${ssmParams.accessToken}`};
        activitiesRes = (await makeRequest(
            "https://www.strava.com/api/v3/athlete/activities",
            {},
            "GET",
            headers
        )) as any[];
    } catch (e) {
        return;
    }

    const types = ["Run", "Hike", "Walk"];
    const badShoes = JSON.parse(process.env.GEAR_IDS as string);

    const needFixing = activitiesRes.filter((act) => {
        if (!types.includes(act.type)) {
            return false;
        }
        return !act.gear_id || badShoes.includes(act.gear_id);
    });

    needFixing.forEach((act) => {
        console.log(act.name, act.type, act.gear_id);
    });

    if (needFixing.length) {
        const Message = needFixing
            .map((a) => `${a.name} (${a.type}) https://www.strava.com/activities/${a.id}`)
            .join("\n");

        const params = {
            Message,
            Subject: "Strava activities with missing footwear",
            TopicArn: process.env.TOPIC_ARN,
        };
        await publishSns(params);
    } else {
        console.info("No missing footwear");
    }
};
