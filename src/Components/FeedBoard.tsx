import { useEffect, useState } from "react";

const MAXKILLMAILS = 10;

type Zkb = {
    hash: string;
    fittedValue: number;
    droppedValue: number;
    destroyedValue: number;
    totalValue: number;
    npc: boolean;
    solo: boolean;
};

type ZKillmail = {
    killmail_id: number;
    zkb: Zkb;
};

type EsiKillmail = {
    attackers: {
        final_blow: boolean;
        character_id: number;
        ship_type_id: number;
        corporation_id: number;
    }[];
    victim: {
        character_id: number;
        corporation_id: number;
        ship_type_id: number;
    };
};

type Character = {
    id: number;
    name: string;
    corporation_id: number;
};

type Killmail = {
    attacker: Character;
    victim: Character;
    zKillmail: ZKillmail;
    esiKillmail: EsiKillmail;
};

type NameProps = {
    name: string;
    pilotId: number;
};

function Name({ name, pilotId }: NameProps) {
    return (
        <div className="py-1 text-ellipsis overflow-hidden text-center text-dark-50 whitespace-nowrap">
            <a
                href={`https://zkillboard.com/character/${pilotId}`}
                className="hover:underline"
            >
                {name}
            </a>
        </div>
    );
}

type ImageProps = {
    killmail: Killmail;
};

const formatter = Intl.NumberFormat("en", {
    notation: "compact",
    maximumFractionDigits: 2,
});

function Image({ killmail }: ImageProps) {
    return (
        <div
            className={
                killmail.attacker.corporation_id === 98684728
                    ? "px-5 pt-5 bg-green-900 rounded"
                    : "px-5 pt-5 bg-red-900 rounded"
            }
        >
            <a
                href={`https://zkillboard.com/kill/${killmail.zKillmail.killmail_id}`}
                className="mx-auto w-32 h-32"
            >
                <img
                    className="rounded"
                    src={`https://images.evetech.net/types/${killmail.esiKillmail.victim.ship_type_id}/render?size=128`}
                    alt="Image of a ship"
                />
            </a>
            <div className="pt text-white text-center text-ellipsis whitespace-nowrap overflow-hidden">
                {`${formatter.format(killmail.zKillmail.zkb.totalValue)} ISK`}
            </div>
        </div>
    );
}

type FeedProps = {
    killmail: Killmail;
};

function Feed({ killmail }: FeedProps) {
    return (
        <div className="rounded w-40 border-gray-600 flex-grow-0 flex-shrink-0">
            <Name
                name={killmail.attacker.name}
                pilotId={killmail.attacker.id}
            />
            <Image killmail={killmail} />
            <Name name={killmail.victim.name} pilotId={killmail.victim.id} />
        </div>
    );
}

export default function FeedBoard() {
    const [kills, setKills] = useState<Killmail[]>([]);

    const fetchCorpZKill = async () => {
        const response = await fetch(
            "https://zkillboard.com/api/corporationID/98684728/"
        );
        if (response.ok) {
            const zKillmails: ZKillmail[] = await response.json();
            let validKills = 0;
            return Promise.all(
                zKillmails
                    .filter((zk) => {
                        if (validKills >= MAXKILLMAILS) {
                            return false;
                        }
                        if (!zk.zkb.npc) {
                            validKills += 1;
                            return true;
                        }
                        return false;
                    })
                    .map(async (zk) => {
                        const esiKillmail = await fetchEsiKillmail(
                            zk.killmail_id,
                            zk.zkb.hash
                        );
                        if (esiKillmail) {
                            const victim = await fetchCharacter(
                                esiKillmail.victim.character_id
                            );
                            const attacker = await fetchCharacter(
                                esiKillmail.attackers.find((c) => {
                                    return c.final_blow;
                                })?.character_id ?? 95421926
                            );
                            if (victim && attacker) {
                                const killmail: Killmail = {
                                    attacker: attacker,
                                    victim: victim,
                                    esiKillmail: esiKillmail,
                                    zKillmail: zk,
                                };
                                return killmail;
                            }
                        }
                        return undefined;
                    })
            );
        }
    };

    const fetchEsiKillmail = async (
        killmailId: number,
        hash: string
    ): Promise<EsiKillmail | undefined> => {
        const response = await fetch(
            `https://esi.evetech.net/latest/killmails/${killmailId}/${hash}/?datasource=tranquility`
        );
        if (response.ok) {
            return await response.json();
        }
    };

    const fetchCharacter = async (
        characterId: number
    ): Promise<Character | undefined> => {
        const response = await fetch(
            `https://esi.evetech.net/latest/characters/${characterId}/?datasource=tranquility`
        );
        if (response.ok) {
            let char: Character = await response.json();
            char.id = characterId;
            return char;
        }
        return undefined;
    };

    useEffect(() => {
        fetchCorpZKill().then((resp) => {
            if (resp) {
                setKills(
                    resp.filter((km) => {
                        return km;
                    }) as Killmail[]
                );
            }
        });
    }, []);

    return (
    kills.length > 0 ? (
            <div className="mx-16 my-4 overflow-x-hidden relative">
            <div className="text-3xl md:text-4xl font-bold text-dark-50 pb-4">
            <a
            className="hover:underline"
            href="https://zkillboard.com/corporation/98684728/"
            >
            Fifty Fifty Fifty Feedboard
            </a>
            </div>
            <div className="text-xl md:text-2xl font-bold text-dark-100 pb-4">
            Displaying last {MAXKILLMAILS} kills
            </div>
            <div className="flex flex-nowrap gap-8 overflow-x-auto p-4">
            {kills.map((k) => {
                    return <Feed killmail={k} key={k.zKillmail.killmail_id} />;
                    })}
            </div>
            <div className="bg-[url('https://images.evetech.net/corporations/98684728/logo?size=256')] bg-no-repeat bg-center bg-contain w-full h-full absolute opacity-10 top-0 left-0 -z-50"></div>
            </div>

    ): (<></>)
    );
}
