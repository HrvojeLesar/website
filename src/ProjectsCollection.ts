import { IInfoCard } from "./IInfoCard";

import PATHFINDER from "./Images/pathfinder.png";

export const projects: IInfoCard[] = [];
export const eveStuff: IInfoCard[] = [
    {
        title: "Pathfinder",
        description: "Pathfinder is a mapping tool for Eve Online.",
        moreInfo:
            "Official Pathfinder website is down and most others requires \
            you to be a part of their alliance or corporation, \
            so I decided to host my own. As is tradition this requires \
            you to be a port of Fifty Fifty Fifty corporation.",
        image: {
            src: PATHFINDER,
            alt: "Pathfinder logo",
        },
        link: "https://pathfinder.hrveklesarov.com",
    },
];
