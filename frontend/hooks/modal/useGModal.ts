import { useState, useCallback } from "react";

export interface ModalContent {
  title?: string;
  content: React.ReactNode;
  size?: "sm" | "md" | "lg" | "xl" | "full";
  onClose?: () => void;
}

export const useGModal = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [modalContent, setModalContent] = useState<ModalContent | null>(null);

  const openModal = useCallback((content: ModalContent) => {
    setModalContent(content);
    setIsOpen(true);
  }, []);

  const closeModal = useCallback(() => {
    setIsOpen(false);
    if (modalContent?.onClose) {
      modalContent.onClose();
    }
    // Clear content after a short delay to allow for smooth closing animation
    setTimeout(() => {
      setModalContent(null);
    }, 150);
  }, [modalContent]);

  const updateModalContent = useCallback((content: Partial<ModalContent>) => {
    if (modalContent) {
      setModalContent({ ...modalContent, ...content });
    }
  }, [modalContent]);

  return {
    isOpen,
    modalContent,
    openModal,
    closeModal,
    updateModalContent,
  };
};

export type ModalHandle = ReturnType<typeof useGModal>;





