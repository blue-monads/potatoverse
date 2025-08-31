"use client";
import React, { useState } from 'react';
import { Search, Filter, MoreHorizontal, UserIcon, Mail, Calendar, Shield, Eye, Edit, Trash2, UserCheck, UserX, LockIcon } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { getUsers, User } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { FantasticTable } from '@/contain';
import { useGApp } from '@/hooks';
import { ColumnDef } from '@/contain/compo/FantasticTable/FantasticTable';
import WithTabbedUserLayout from './WithTabbedUserLayout';


const columns = [
  {
    title: '#',
    key: 'id',
    render: (cellData: any, row: User) => {

      console.log("row", row);

      return <div>
        <img src={`/z/profileImage/${row.id}/${(row.name)}`} alt="profile" className="w-8 h-8 rounded-full" />

      </div>
    },
  },
  {
    title: 'Name',
    key: 'name',
  },
  {
    title: 'Email',
    key: 'email',
  },
  {
    title: 'Utype',
    key: 'utype',
  },
  {
    title: 'Disabled',
    key: 'disabled',
    render: "boolean",
  },

  {
    title: 'Created At',
    key: 'createdAt',
  },


] as ColumnDef[];

export default function Page() {
  const [searchTerm, setSearchTerm] = useState('');
  const gapp = useGApp();

  const loader = useSimpleDataLoader<User[]>({
    loader: getUsers,
    ready: gapp.isInitialized,
  });


  console.log("loader", loader);



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
        placeholder="Search users..."
      />

      <WithTabbedUserLayout activeTab="users">
        <FantasticTable
          isLoading={loader.loading}
          classNamesContainer='w-full p-1 max-w-7xl mx-auto'
          classNamesTable='border border-primary-50 rounded-md'
          classNamesTableHead='uppercase'
          columns={columns}
          data={loader.data || []}
          actions={[
            {
              label: "View",
              className: "bg-primary-500",
              onClick: (rowData: User) => {
                console.log("rowData", rowData);
              },
              icon: <Eye className="w-4 h-4" />,
            },

            {
              label: "Edit",
              onClick: (rowData: User) => {
                console.log("rowData", rowData);
              },
              className: "bg-secondary-500",
              icon: <Edit className="w-4 h-4" />,
            },

            {
              label: "Change Password",
              dropdown: true,
              onClick: (rowData: User) => {
                console.log("rowData", rowData);
              },
              icon: <LockIcon className="w-4 h-4" />,
            },

            {
              label: "Change Status",
              dropdown: true,
              onClick: (rowData: User) => {
                console.log("rowData", rowData);
              },
              icon: <UserX className="w-4 h-4" />,
            },


            {
              label: "Delete",
              dropdown: true,
              onClick: (rowData: User) => {
                console.log("rowData", rowData);
              },
              icon: <Trash2 className="w-4 h-4" />,
            },

          ]}

        />
      </WithTabbedUserLayout>



    </WithAdminBodyLayout>
  </>)
}


