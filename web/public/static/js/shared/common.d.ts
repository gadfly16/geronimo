export interface msg {
    Kind: number;
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
    Create = 7,
    AuthUser = 8,
    GetTree = 9,
    Tree = 10,
    GetCopy = 11,
    GetDisplay = 12,
    Display = 13
}
export declare enum WSMsg {
    Credentials = 0,
    Subscribe = 1,
    Unsubscribe = 2,
    Update = 3,
    Error = 4,
    ClientShutdown = 5,
    Heartbeat = 6
}
export declare enum nodeKinds {
    Root = 0,
    Group = 1,
    User = 2,
    Broker = 3,
    Account = 4
}
export declare let NodeKindName: string[];
