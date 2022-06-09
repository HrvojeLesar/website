import { CardCollectionProps } from "./CardCollection";
import { other, projects } from "./ProjectsCollection";

export const cardCollections: CardCollectionProps[] = [
    {
        title: "Projects",
        cards: projects,
    },
    {
        title: "Other",
        cards: other,
    },
];
