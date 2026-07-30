# Overview

## Management Hierarchy

The first hierarchy is a hierarchy of groups akin to an org-chart. Going up the hierarchy corresponds to going up in power. Upper levels have management power over lower levels.


```
                     admin
                  /         \
                 /           \
 staff_management            user_groups
    /         \                /     \
m1_mgers     m2_mgers     club1     club2     
    |           |
m1_staff     m2_staff
```

In this example, m1_mgers {Makerspace 1 Managers} has the ability to manage m1_staff {Makerspace 1 Staff}. This means users in the m1_mgers group can add and remove users to the m1_staff group but cannot modify their own group. 


## Membership
The first hierarchy describes how groups manage each down the org chart. Importantly, membership does not travel up the tree. In example 1, m1_mgers are _managed by_ staff_management but they are _not_ _members of_ of the staff_management group.




### Permissions
When a user is a direct member of a group, there are options for their visibility into the group. From most-restrictive to least restrictive

1. see-none: user does not even know they are in the group
2. see-self: user knows that the are in the group, but don't know anyone else in the group
3. see-all: user can see that they are in the group as well as everyone else in the group

## Subgroup Hierarchy

The second hieararchy of groups is that of supergroup:subgroup relations sometimes called union or combination groups. If we wanted to be able to refer to all staff, we could create a combination of m1_staff and m2_staff called all_staff. To do this, we can add m1_staff and m2_staff as subgroups of the supergroup all_staff. Here, subgroup relationships are shown with `+` connections. 


```           
     staff_management
    /         |       \
m1_mgers    staff    m2_mgers
    |       +   +       |
m1_staff>++++   ++++<m2_staff
```


When a subgroup relation is created, it carries a permission level like direct membership. If a user is a member of multiple direct subgroups of a supergroup, the highest permission on the group_subgroup relation is used for that users permissions within their supergroup. 




## Group Visibility


# Implementation

## Tables

### groups

```
id, name, manager_id
```

### group_direct_membership
```
group_id, user_id, permission
```

### group_direct_subgroups
```
group_id, subgroup_id, permission
```


### group_direct_shares
```
shared_group_id, shared_to_group_id
```


## Views

### group_management
```
manager_group_id, managee_group_id
```

Shows if manager_group_id can manage managee_group_id. This includes all levels of the hierarchy under manager_group_id rather than just direct managees, 


### group_visibility
```
observing_group_id,visible_group_id
```

### group_subgroups
```
group_id,subgroup_id,permission
```
permission comes from direct subgrouping closest to top group_id

### group_membership
```
group_id, user_id, permission
```
union of direct membership or subgroup membership

