"use client"
import { AddButton } from "@/contain/AddButton";
import WithAdminBodyLayout from "@/contain/Layouts/WithAdminBodyLayout";
import { Mail, UserIcon } from "lucide-react";
import WithTabbedUserLayout from "../WithTabbedUserLayout";
import BigSearchBar from "@/contain/compo/BigSearchBar";
import { useState } from "react";

export default function Page() {
    const [searchTerm, setSearchTerm] = useState('');
    return (<>

        <WithAdminBodyLayout
            Icon={UserIcon}
            name='Users'
            description="Manage your users, roles, and permissions."
            rightContent={<>
                <AddButton
                    name="+ User"
                    onClick={() => { }}
                />

            </>}

        >

            <BigSearchBar
                setSearchText={setSearchTerm}
                searchText={searchTerm}
                placeholder="Search invites..."
            />

            <WithTabbedUserLayout activeTab="invites">
                <div className="max-w-7xl mx-auto">
                    "No invites yet"

                </div>
            </WithTabbedUserLayout>


        </WithAdminBodyLayout>


    </>)
}