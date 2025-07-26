

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


                <div className="tx-card tx-card-bordered p-4 bg-primary-hover-token w-full max-w-lg">
                    <div className="tx-card-title">Button</div>
                    <div className="tx-card-body flex gap-2 items-start">
                        <button className="tx-btn tx-btn-primary">Primary Button</button>
                        <button className="tx-btn tx-btn-secondary">Secondary Button</button>
                        <button className="tx-btn tx-btn-accent">Accent Button</button>
                    </div>
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

            <div className="tx-card p-4">




                <div className="tx-nav">

                    <div className="tx-nav-item !text-[var(--color-text)]">Logo</div>

                    <div className="tx-nav-items">
                        <div className="tx-nav-item">Home</div>
                        <div className="tx-nav-item">About</div>
                        <div className="tx-nav-item">Services</div>
                        <div className="tx-nav-item">Contact</div>

                    </div>


                </div>

                {/* LOGIN form card */}

                <div className="tx-card tx-card-bordered p-4 bg-accent-hover-token w-full max-w-md">
                    <div className="tx-card-title">Login Form</div>
                    <div className="tx-card-body flex flex-col gap-4">
                        <input type="text" placeholder="Username" className="tx-input" />
                        <input type="password" placeholder="Password" className="tx-input" />
                        <button className="tx-btn tx-btn-primary">Login</button>
                    </div>
                    <div className="tx-card-footer flex flex-col items-center">
                        <button className="tx-btn tx-btn-secondary">Forgot Password?</button>
                        <button className="tx-btn tx-btn-link">Sign Up</button>
                    </div>
                </div>

                <div className="flex gap-2">
                    <div className="tx-badge tx-badge-primary">Primary</div>
                    <div className="tx-badge tx-badge-secondary">Secondary</div>
                    <div className="tx-badge tx-badge-accent">Accent</div>
                    <div className="tx-badge tx-badge-danger">Danger</div>
                    <div className="tx-badge tx-badge-success">Success</div>
                    <div className="tx-badge tx-badge-warning">Warning</div>
                    <div className="tx-badge tx-badge-info">Info</div>
                    <div className="tx-badge tx-badge-outline">Outline</div>
                    <div className="tx-badge tx-badge-ghost">Ghost</div>
                    <div className="tx-badge tx-badge-link">Link</div>
                    <div className="tx-badge tx-badge-close">Close</div>
                </div>

                {/* Table showcase */}
                <div className="tx-card tx-card-bordered p-4 bg-secondary-hover-token w-full max-w-2xl text-xs">
                    <div className="tx-card-title">Data Table</div>
                    <div className="tx-card-body">
                        <table className="tx-table w-full">
                            <thead>
                                <tr className="tx-table-header">
                                    <th className="tx-table-cell">ID</th>
                                    <th className="tx-table-cell">Name</th>
                                    <th className="tx-table-cell">Email</th>
                                    <th className="tx-table-cell">Role</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr className="tx-table-row">
                                    <td className="tx-table-cell">1</td>
                                    <td className="tx-table-cell">John Doe</td>
                                    <td className="tx-table-cell">
                                        <input type="email" className="tx-input" placeholder="Email" />
                                    </td>
                                    <td className="tx-table-cell">
                                        <select className="tx-select">
                                            <option value="admin">Admin</option>
                                            <option value="user">User</option>
                                        </select>
                                    </td>
                                </tr>
                                <tr className="tx-table-row">
                                    <td className="tx-table-cell">2</td>
                                    <td className="tx-table-cell">Jane Smith</td>
                                    <td className="tx-table-cell">
                                        <input type="email" className="tx-input" placeholder="Email" />
                                    </td>
                                    <td className="tx-table-cell">
                                        <select className="tx-select">
                                            <option value="admin">Admin</option>
                                            <option value="user">User</option>
                                        </select>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>



            </div>
        </div>
    );
}





export default Page;