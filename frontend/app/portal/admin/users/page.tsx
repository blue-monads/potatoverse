"use client";
import React, { useState } from 'react';
import { Search, Filter, MoreHorizontal,  UserIcon, Mail, Calendar, Shield, Eye, Edit, Trash2, UserCheck, UserX } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { getUsers, User } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { FantasticTable } from '@/contain';
import { useGApp } from '@/hooks';
import { ColumnDef } from '@/contain/compo/FantasticTable/FantasticTable';


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

  {
    title: 'Actions',
    key: 'actions',
  },


] as ColumnDef[];

export default function Page() {
  const [searchTerm, setSearchTerm] = useState('');
  const gapp = useGApp();
  
  const loader  = useSimpleDataLoader<User[]>({
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
      />

      <FantasticTable
        isLoading={loader.loading}
        classNamesContainer='w-full p-1 max-w-7xl mx-auto'
        classNamesTable='border border-primary-50 rounded-md'
        classNamesTableHead='uppercase'
        columns={columns}
        data={loader.data || []}

      />



    </WithAdminBodyLayout>
  </>)
}


