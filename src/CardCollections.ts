import { CardCollectionProps } from "./CardCollection";
import { eveStuff, projects } from "./ProjectsCollection";

export const cardCollections: CardCollectionProps[] = [
    {
        title: "Projects",
        cards: projects,
    },
    {
        title: "Eve stuff",
        cards: eveStuff,
    },
];
