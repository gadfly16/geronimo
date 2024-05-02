export interface GuiMessage {
    Type: string;
    Payload: any;
}
export declare let guiMessageType: {
    getUserTree: string;
    userTree: string;
};
export declare enum NodeType {
    Root = 0,
    User = 1,
    Account = 2,
    Broker = 3,
    Group = 4,
    Pocket = 5
}
export declare let NodeTypeName: string[];
