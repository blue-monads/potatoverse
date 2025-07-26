

/*

one page showcase of all my custom comeponents

tx-btn
tx-btn-sm
tx-btn-lg
tx-btn-primary
tx-btn-secondary
tx-btn-accent
tx-btn-ghost
tx-btn-link
tx-btn-close

tx-card
tx-card-bordered
tx-card-compact
tx-card-body
tx-card-title

tx-grid # mobile-1, desktop-4, md-3


tx-input
tx-textarea
tx-select
tx-checkbox
tx-radio
tx-switch

tx-modal
tx-modal-header
tx-modal-body
tx-modal-footer

tx-tab
tx-tab-list
tx-tab-panel
tx-tab-content

tx-table
tx-table-header
tx-table-row
tx-table-cell
tx-table-footer



*/


const Page = () => {
    return (
        <div className="flex h-full w-full items-center justify-center">
            <div className="flex flex-col items-center justify-center space-y-4">
                <h1 className="text-3xl font-bold">Theme Showcase</h1>
                <p className="text-lg text-gray-700">Explore the custom components and styles</p>

                <div className="flex gap-2">


                    {/* make a card and inside card add other component */}

                    <div className="tx-card tx-card-bordered p-4 bg-primary-hover-token">
                        <div className="tx-card-title">Button</div>
                        <div className="tx-card-body flex gap-2 items-start">                            
                            <button className="tx-btn tx-btn-primary">Primary Button</button>
                            <button className="tx-btn tx-btn-secondary">Secondary Button</button>
                            <button className="tx-btn tx-btn-accent">Accent Button</button>
                        </div>
                    </div>

                    <div className="tx-card tx-card-bordered p-4 bg-secondary-hover-token">
                        <div className="tx-card-title">Input</div>
                        <div className="tx-card-body flex flex-col gap-2">
                            <input type="text" placeholder="Text Input" className="tx-input" />
                            <textarea placeholder="Textarea" className="tx-textarea"></textarea>
                            <select className="tx-select">
                                <option>Select Option</option>
                                <option>Option 1</option>
                                <option>Option 2</option>
                            </select>
                        </div>
                        <div className="tx-card-footer">
                            <button className="tx-btn tx-btn-primary">Submit</button>
                        </div>
                    </div>

                </div>

            </div>
           


        </div>
    );
}





export default Page;