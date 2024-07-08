export interface msg {
    Kind: Number;
    Payload: any;
}
export declare enum msgKinds {
    OK = 0,
    Error = 1,
    Stop = 2,
    Stopped = 3,
    Update = 4,
    Parms = 5,
    GetParms = 6,
    Create = 7
}
export declare let WSMsg: {
    Credentials: string;
    Subscribe: string;
    Unsubscribe: string;
    Update: string;
};
export declare enum nodeKinds {
    Root = 0,
    Group = 1,
    User = 2,
    Broker = 3,
    Account = 4
}
export declare enum payloadKinds {
    UserNodePayload = 0
}
export declare let NodeTypeName: string[];
