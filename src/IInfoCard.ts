import { ReactNode } from "react";

export interface IInfoCard {
    title: string;
    description: string | ReactNode;
    moreInfo?: string | ReactNode;
    image: {
        src: string | ReactNode;
        alt: string;
    };
    link?: string;
}
