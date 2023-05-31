import {GetParametersByPathCommand, Parameter, PutParameterCommand, SSMClient} from "@aws-sdk/client-ssm";

export const SSM_PREFIX = "/strava/";

export class ParamStore {
    private ssm: SSMClient;

    constructor(ssm: SSMClient) {
        this.ssm = ssm;
    }

    public getParams(): Promise<StravaParams> {
        const copy = this;

        const params = {
            Path: SSM_PREFIX,
            Recursive: true,
            WithDecryption: true,
        };

        return this.ssm.send(new GetParametersByPathCommand(params)).then((data) => {
            return (data.Parameters as Parameter[]).reduce((acc: any, val: any) => {
                const key = val.Name.replace(SSM_PREFIX, "");
                acc[key] = val.Value;
                return acc;
            }, {});
        });
    }

    public setParam(param: keyof StravaParams, value: string, type = "String"): Promise<void> {
        const data = {
            Name: `${SSM_PREFIX}${param}`,
            Value: value,
            Type: type,
            Overwrite: true,
        };

        return this.ssm.send(new PutParameterCommand(data)).then();
    }
}

export type StravaParams = {
    clientId: string;
    clientSecret: string;
    accessToken: string;
    refreshToken: string;
};
