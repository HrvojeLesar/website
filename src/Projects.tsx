import InfoCard from "./InfoCard";

export default function Projects() {
    return (
        <>
            <div className="text-5xl md:text-6xl font-bold text-dark-50 px-16 py-16">
                Projects
            </div>
            <InfoCard
                title="Svetski naslov"
                content="Svetski content"
            />
            <InfoCard
                title="Svetski naslov"
                content="Svetski content"
                reverse
            />
        </>
    );
}
