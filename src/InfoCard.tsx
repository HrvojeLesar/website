import { ReactNode } from "react";

type InfoCardProps = {
    title: string;
    content: string | ReactNode;
    reverse?: boolean;
};

export default function InfoCard({ title, reverse, content }: InfoCardProps) {
    console.log(typeof content);
    return (
        <div
            className={`grid lg:grid-cols-2 px-16 py-24 gap-12 ${
                reverse ? "bg-dark-400" : "bg-dark-500"
            }`}
        >
            <a
                className={`${reverse ? "order-last" : "order-first"} mx-auto`}
                href={`https://zkillboard.com/kill/101297554`}
            >
                <img
                    className="rounded w-32 h-32"
                    src={`https://images.evetech.net/types/670/render?size=128`}
                    alt="Image of a ship"
                />
            </a>
            <div className="flex flex-col justify-center items-center">
                <div>
                    <div className="text-3xl md:text-4xl font-bold text-dark-50">
                        {title}
                    </div>
                    {typeof content === "string" ? (
                        <div className="text-xl text-dark-50">{content}</div>
                    ) : (
                        content
                    )}
                </div>
            </div>
        </div>
    );
}
