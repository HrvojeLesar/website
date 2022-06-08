import { IInfoCard } from "./IInfoCard";
import InfoCard from "./InfoCard";

type CardCollectionProps = {
    title?: string;
    cards: IInfoCard[];
};

export default function CardCollection({ title, cards }: CardCollectionProps) {
    return (
        <div className="overflow-x-hidden overflow-y-auto">
            <div className="text-5xl md:text-6xl font-bold text-dark-50 px-16 py-16">
                {title ?? ""}
            </div>
            {cards.map((c, index) => {
                return (
                    <InfoCard
                        key={c.title}
                        infoCard={c}
                        reverse={index % 2 === 0}
                    />
                );
            })}
        </div>
    );
}
