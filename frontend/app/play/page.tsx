"use client"
import { useGApp } from "@/hooks";
import ModalDemo from "@/hooks/modal/use";

export default function PlayPage() {
  const { modal } = useGApp();

  return (
        <ModalDemo />
  );
}
