import { ReactNode } from "react";
import { IconType } from "react-icons";

export interface IInfoCard {
    title: string;
    description: string | ReactNode;
    moreInfo?: string | ReactNode;
    note?: string;
    image: {
        src: string | ReactNode;
        alt: string;
    };
    link?: string;
    linkButtons?: {
        title: string;
        url: string;
        icon: IconType;
        iconDecsShort: string;
    }[];
}
