import { IInfoCard } from "./IInfoCard";

type InfoCardProps = {
    infoCard: IInfoCard;
    reverse?: boolean;
};

export default function InfoCard({ infoCard, reverse }: InfoCardProps) {
    return (
        <div className={reverse ? "bg-dark-500" : "bg-dark-400"}>
            <div
                className={`grid lg:grid-cols-2 px-16 py-24 gap-12 md:mx-32 sm:mx-0`}
            >
                <a
                    className={`${
                        reverse ? "order-last" : "lg:order-first order-last"
                    } mx-auto self-center`}
                    href={infoCard.link ?? `#${infoCard.title}`}
                >
                    {typeof infoCard.image.src === "string" ? (
                        <img
                            className="rounded"
                            src={infoCard.image.src}
                            alt={infoCard.image.alt}
                        />
                    ) : (
                        infoCard.image.src
                    )}
                </a>
                <div className="flex flex-col justify-center items-center">
                    <div className="self-start">
                        <div
                            id={infoCard.link ? undefined : infoCard.title}
                            className="text-3xl md:text-4xl font-bold text-dark-50 pb-4"
                        >
                            <a
                                className="hover:underline"
                                href={infoCard.link ?? `#${infoCard.title}`}
                            >
                                {infoCard.title}
                            </a>
                        </div>
                        {typeof infoCard.description === "string" ? (
                            <div className="leading-relaxed text-xl text-dark-50">
                                {infoCard.description}
                            </div>
                        ) : (
                            infoCard.description
                        )}
                        {infoCard.moreInfo === undefined ? (
                            <></>
                        ) : (
                            <div className="pt-1 leading-relaxed text-lime-lg text-dark-200">
                                {infoCard.moreInfo}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}
