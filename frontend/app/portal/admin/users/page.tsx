"use client";
import React, { useState } from 'react';
import { Search, Filter, MoreHorizontal, User, Mail, Calendar, Shield, Eye, Edit, Trash2, UserCheck, UserX } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { getUsers } from '@/lib';

export default function Page() {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedUsers, setSelectedUsers] = useState<number[]>([]);
 

  return (<>
    <WithAdminBodyLayout
      Icon={User}
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

      <div>TODO</div>

    </WithAdminBodyLayout>
  </>)
}


