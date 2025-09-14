"use client"
import { ModalHandle, useGApp } from "@/hooks";

export default function ModalDemo() {
  const { modal } = useGApp();

  const showSimpleModal = () => {
    modal.openModal({
      title: "Simple Modal",
      content: (
        <div className="text-center">
          <p className="text-gray-600 dark:text-gray-300 mb-4">
            This is a simple modal with a title and content.
          </p>
          <button
            onClick={() => modal.closeModal()}
            className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors"
          >
            Close Modal
          </button>
        </div>
      ),
      size: "md"
    });
  };

  const showLargeModal = () => {
    modal.openModal({
      title: "Large Modal",
      content: (
        <div className="space-y-4">
          <p className="text-gray-600 dark:text-gray-300">
            This is a large modal that can contain more content.
          </p>
          <div className="bg-gray-100 dark:bg-gray-700 p-4 rounded-lg">
            <h3 className="font-semibold mb-2">Sample Content</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              You can put any React components here, including forms, tables, or other complex UI elements.
            </p>
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => modal.closeModal()}
              className="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={() => modal.closeModal()}
              className="bg-primary-500 hover:bg-primary-600 text-white px-4 py-2 rounded-lg transition-colors"
            >
              Confirm
            </button>
          </div>
        </div>
      ),
      size: "lg"
    });
  };

  const showFullScreenModal = () => {
    modal.openModal({
      title: "Full Screen Modal",
      content: <BigOne modal={modal} />,
      size: "full"
    });
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-8 text-center">
          Modal Demo
        </h1>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-lg">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">Simple Modal</h2>
            <p className="text-gray-600 dark:text-gray-300 mb-4">
              A basic modal with title and simple content.
            </p>
            <button
              onClick={showSimpleModal}
              className="w-full bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg transition-colors"
            >
              Open Simple Modal
            </button>
          </div>

          <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-lg">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">Large Modal</h2>
            <p className="text-gray-600 dark:text-gray-300 mb-4">
              A larger modal for more complex content and forms.
            </p>
            <button
              onClick={showLargeModal}
              className="w-full bg-primary-500 hover:bg-primary-600 text-white px-4 py-2 rounded-lg transition-colors"
            >
              Open Large Modal
            </button>
          </div>

          <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-lg">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">Full Screen</h2>
            <p className="text-gray-600 dark:text-gray-300 mb-4">
              A full-screen modal for complex dashboards or forms.
            </p>
            <button
              onClick={showFullScreenModal}
              className="w-full bg-purple-500 hover:bg-purple-600 text-white px-4 py-2 rounded-lg transition-colors"
            >
              Open Full Screen Modal
            </button>
          </div>
        </div>

        <div className="mt-12 bg-white dark:bg-gray-800 p-6 rounded-lg shadow-lg">
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-4">How to Use</h2>
          <div className="space-y-4 text-gray-600 dark:text-gray-300">
            <p>
              The modal system is now integrated into the global app state. You can access it from anywhere in your app using:
            </p>
            <pre className="bg-gray-100 dark:bg-gray-700 p-4 rounded-lg overflow-x-auto">
              <code>{`const { modal } = useGApp();

// Open a modal
modal.openModal({
  title: "Modal Title",
  content: <YourContent />,
  size: "md", // sm, md, lg, xl, full
  onClose: () => console.log("Modal closed")
});

// Close the modal
modal.closeModal();`}</code>
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
}


const BigOne = ({ modal }: { modal: ModalHandle }) => {
  return (<>(
        <div className="h-full flex flex-col">
          <div className="flex-1 bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-800 dark:to-gray-900 p-6 rounded-lg">
            <h3 className="text-xl font-semibold mb-4">Full Screen Content</h3>
            <p className="text-gray-600 dark:text-gray-300 mb-4">
              This modal takes up the full screen and is perfect for complex forms or dashboards.
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="bg-white dark:bg-gray-700 p-4 rounded-lg">
                <h4 className="font-medium mb-2">Section 1</h4>
                <p className="text-sm text-gray-600 dark:text-gray-400">Content for section 1</p>
              </div>
              <div className="bg-white dark:bg-gray-700 p-4 rounded-lg">
                <h4 className="font-medium mb-2">Section 2</h4>
                <p className="text-sm text-gray-600 dark:text-gray-400">Content for section 2</p>
              </div>
            </div>
          </div>
          <div className="mt-4 flex justify-end">
            <button
              onClick={() => modal.closeModal()}
              className="bg-blue-500 hover:bg-blue-600 text-white px-6 py-2 rounded-lg transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      )
  
  </>)
}
