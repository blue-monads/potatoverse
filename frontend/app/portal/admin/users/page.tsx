"use client";
import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Search, Filter, MoreHorizontal, UserIcon, Mail, Calendar, Shield, Eye, Edit, Trash2, UserCheck, UserX, LockIcon, MailIcon } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { getUsers, User } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { FantasticTable } from '@/contain';
import { ModalHandle, useGApp } from '@/hooks';
import { ColumnDef } from '@/contain/compo/FantasticTable/FantasticTable';
import WithTabbedUserLayout from './WithTabbedUserLayout';
import AddInviteModal from './sub/AddInviteModal';


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
  const router = useRouter();
  const gapp = useGApp();

  const loader = useSimpleDataLoader<User[]>({
    loader: getUsers,
    ready: gapp.isInitialized,
  });

  const handleUserAdded = () => {
    loader.reload();
  };

  const handleCreateUser = () => {
    router.push('/portal/admin/users/create');
  };

  console.log("loader", loader);



  return (<>
    <WithAdminBodyLayout
      Icon={UserIcon}
      name='Users'
      description="Manage your users, roles, and permissions."
      rightContent={<>
        <button
          onClick={() => showInviteUserModal(gapp.modal, handleUserAdded)}
          className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:text-gray-800 dark:text-gray-300 dark:hover:text-white transition-colors"
        >
          <MailIcon className="w-4 h-4" />
          Invite User
        </button>
        <AddButton
          name="+ User"
          onClick={handleCreateUser}
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


const showInviteUserModal = (modal: ModalHandle, onUserAdded: () => void) => {
  modal.openModal({
    title: "Invite User",
    content: <AddInviteModal onInviteAdded={onUserAdded} />,
    size: "lg"
  });
};
