import { CardCollectionProps } from "../Components/CardCollection";
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
