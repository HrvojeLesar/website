import { ReactNode } from "react";
import { IconType } from "react-icons";

type NavbarButtonProps = {
    icon?: IconType;
    value: string;
    url: string;
};


export function NavbarButton({ value, url, icon }: NavbarButtonProps) {
    return (
    <>
        <a
            className="hover:bg-dark-500 rounded text-white p-3 box-content"
            href={url}
        >
        {icon ? (
                <div className="flex gap-2 justify-center items-center">
                {icon ? (icon({size: 28})) : (<></>)}
                {value}
                </div>
        ): (value)}
        </a>
    </>
    );
}

type NavbarProps = {
    buttonsLeft?: ReactNode;
    buttonsRight?: ReactNode;
};

export default function Navbar({ buttonsLeft, buttonsRight }: NavbarProps) {
    return (
        <div className="sticky top-0 left-0 flex justify-between p-2 bg-black opacity-70">
            {buttonsLeft ? (
                <div className="flex gap-2">{buttonsLeft}</div>
            ) : (
                <></>
            )}
            {buttonsRight ? (
                <div className="flex gap-2">{buttonsRight}</div>
            ) : (
                <></>
            )}
        </div>
    );
}